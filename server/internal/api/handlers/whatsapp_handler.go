package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/lnardon/arete/internal/ai"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/models"
	"github.com/lnardon/arete/internal/repository"
	"github.com/lnardon/arete/internal/whatsapp"
	openai "github.com/openai/openai-go/v3"
)

const (
	activeConversationWindow = 3 * time.Hour
	conversationHistoryLimit = 21 // includes the newest (just-inserted) user turn
	linkCodeTTL              = 10 * time.Minute
	inboundQueueSize         = 64
	webhookWorkerCount       = 4
	messageProcessingTimeout = 60 * time.Second
)

var linkCommandRe = regexp.MustCompile(`(?i)^LINK\s+(\d{6})$`)

type inboundMessage struct {
	PhoneNumber       string
	Text              string
	WhatsAppMessageID string
	RemoteJid         string
	FromMe            bool // part of the media message key; always false here (Webhook already filters FromMe messages)
	IsAudio           bool
	AudioMimetype     string
}

type evolutionWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		Key struct {
			RemoteJid string `json:"remoteJid"`
			FromMe    bool   `json:"fromMe"`
			ID        string `json:"id"`
		} `json:"key"`
		Message struct {
			Conversation        string `json:"conversation"`
			ExtendedTextMessage struct {
				Text string `json:"text"`
			} `json:"extendedTextMessage"`
			AudioMessage *struct {
				URL      string `json:"url"`
				Mimetype string `json:"mimetype"`
				Ptt      bool   `json:"ptt"`
			} `json:"audioMessage,omitempty"`
		} `json:"message"`
	} `json:"data"`
}

type WhatsAppHandler struct {
	whatsappRepo   *repository.WhatsAppRepository
	convRepo       *repository.ConversationRepository
	agent          *ai.Agent
	evoClient      *whatsapp.Client
	webhookSecret  string
	phoneRateLimit middlewareRateLimiter
	queue          chan inboundMessage
}

type middlewareRateLimiter interface {
	Allow(key string) bool
}

func NewWhatsAppHandler(
	whatsappRepo *repository.WhatsAppRepository,
	convRepo *repository.ConversationRepository,
	agent *ai.Agent,
	evoClient *whatsapp.Client,
	webhookSecret string,
	phoneRateLimit middlewareRateLimiter,
) *WhatsAppHandler {
	h := &WhatsAppHandler{
		whatsappRepo:   whatsappRepo,
		convRepo:       convRepo,
		agent:          agent,
		evoClient:      evoClient,
		webhookSecret:  webhookSecret,
		phoneRateLimit: phoneRateLimit,
		queue:          make(chan inboundMessage, inboundQueueSize),
	}

	for range webhookWorkerCount {
		go h.worker()
	}

	return h
}

func (h *WhatsAppHandler) worker() {
	for msg := range h.queue {
		ctx, cancel := context.WithTimeout(context.Background(), messageProcessingTimeout)
		h.processMessage(ctx, msg)
		cancel()
	}
}

// Webhook receives Evolution API's messages.upsert callbacks. It acknowledges
// fast (Evolution retries on anything but a prompt 2xx) and hands real work
// off to the worker pool.
//
// Authenticated via a `token` query parameter rather than a header: Evolution
// API has no mechanism to attach custom headers to outbound webhook calls
// (a feature request for this was explicitly rejected upstream), so the
// shared secret is baked into WEBHOOK_GLOBAL_URL's query string instead.
// This is weaker than a header (the token can end up in Evolution API's own
// access logs) — network isolation, not this token, is the primary defense
// in production; see DEVELOPMENT.md.
func (h *WhatsAppHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if h.webhookSecret == "" || r.URL.Query().Get("token") != h.webhookSecret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload evolutionWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	if payload.Event != "messages.upsert" || payload.Data.Key.FromMe {
		return
	}

	remoteJid := payload.Data.Key.RemoteJid
	if remoteJid == "" || strings.HasSuffix(remoteJid, "@g.us") {
		return // ignore group messages / malformed payloads
	}
	phone := strings.SplitN(remoteJid, "@", 2)[0]

	text := payload.Data.Message.Conversation
	if text == "" {
		text = payload.Data.Message.ExtendedTextMessage.Text
	}
	text = strings.TrimSpace(text)

	audio := payload.Data.Message.AudioMessage
	isAudio := audio != nil
	if text == "" && !isAudio {
		go h.sendBestEffort(phone, "I can only read text or voice messages right now.")
		return
	}

	msg := inboundMessage{
		PhoneNumber:       phone,
		Text:              text,
		WhatsAppMessageID: payload.Data.Key.ID,
		RemoteJid:         remoteJid,
		FromMe:            payload.Data.Key.FromMe,
		IsAudio:           isAudio,
	}
	if isAudio {
		msg.AudioMimetype = audio.Mimetype
	}

	select {
	case h.queue <- msg:
	default:
		slog.Warn("whatsapp processing queue full, handling inline", "phone", phone)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), messageProcessingTimeout)
			defer cancel()
			h.processMessage(ctx, msg)
		}()
	}
}

func (h *WhatsAppHandler) processMessage(ctx context.Context, msg inboundMessage) {
	link, err := h.whatsappRepo.GetLinkByPhone(ctx, msg.PhoneNumber)
	if errors.Is(err, repository.ErrNotFound) {
		h.handleUnlinkedMessage(ctx, msg)
		return
	}
	if err != nil {
		slog.Error("whatsapp: lookup link failed", "error", err)
		return
	}

	if !h.phoneRateLimit.Allow(msg.PhoneNumber) {
		h.sendBestEffort(msg.PhoneNumber, "You're sending messages a bit fast — try again in a minute.")
		return
	}

	userID := link.UserID

	if msg.IsAudio {
		transcript, ok := h.transcribeInbound(ctx, msg)
		if !ok {
			return // transcribeInbound already sent a friendly reply
		}
		msg.Text = transcript
	}

	conv, err := h.convRepo.GetOrCreateActiveConversation(ctx, userID, activeConversationWindow)
	if err != nil {
		slog.Error("whatsapp: get/create conversation failed", "error", err)
		h.sendBestEffort(msg.PhoneNumber, "Sorry, I hit an error — try again in a moment.")
		return
	}

	userContent, err := json.Marshal(map[string]string{"text": msg.Text})
	if err != nil {
		slog.Error("whatsapp: marshal user content failed", "error", err)
		return
	}
	whatsappMessageID := msg.WhatsAppMessageID
	if _, err := h.convRepo.AppendMessage(ctx, conv.ID, "user", userContent, &whatsappMessageID); err != nil {
		if errors.Is(err, repository.ErrDuplicateMessage) {
			return // webhook retry of a message we already fully processed
		}
		slog.Error("whatsapp: append user message failed", "error", err)
		h.sendBestEffort(msg.PhoneNumber, "Sorry, I hit an error — try again in a moment.")
		return
	}

	recent, err := h.convRepo.RecentMessages(ctx, conv.ID, conversationHistoryLimit)
	if err != nil {
		slog.Error("whatsapp: fetch recent messages failed", "error", err)
		h.sendBestEffort(msg.PhoneNumber, "Sorry, I hit an error — try again in a moment.")
		return
	}
	var history []models.AIMessage
	if len(recent) > 0 {
		history = recent[:len(recent)-1] // drop the user turn we just inserted; Agent appends it itself
	}

	reply, err := h.agent.ProcessMessage(ctx, userID, toOpenAIMessages(history), msg.Text)
	if err != nil {
		slog.Error("whatsapp: agent processing failed", "error", err)
		h.sendBestEffort(msg.PhoneNumber, "Sorry, I hit an error processing that — try again in a moment.")
		return
	}

	assistantContent, err := json.Marshal(map[string]string{"text": reply})
	if err != nil {
		slog.Error("whatsapp: marshal assistant content failed", "error", err)
	} else if _, err := h.convRepo.AppendMessage(ctx, conv.ID, "assistant", assistantContent, nil); err != nil {
		slog.Error("whatsapp: append assistant message failed", "error", err)
	}

	h.sendBestEffort(msg.PhoneNumber, reply)
}

func (h *WhatsAppHandler) transcribeInbound(ctx context.Context, msg inboundMessage) (string, bool) {
	media, err := h.evoClient.GetBase64Media(ctx, msg.RemoteJid, msg.WhatsAppMessageID, msg.FromMe)
	if err != nil {
		slog.Error("whatsapp: media download failed", "error", err, "phone", msg.PhoneNumber)
		h.sendBestEffort(msg.PhoneNumber, "Sorry, I couldn't download that voice message — try again or type it instead?")
		return "", false
	}

	audioBytes, err := base64.StdEncoding.DecodeString(media.Base64)
	if err != nil {
		slog.Error("whatsapp: media base64 decode failed", "error", err, "phone", msg.PhoneNumber)
		h.sendBestEffort(msg.PhoneNumber, "Sorry, I couldn't read that voice message — try again or type it instead?")
		return "", false
	}

	filename := "voice." + extensionForMimetype(media.Mimetype, msg.AudioMimetype)
	transcript, err := h.agent.Transcribe(ctx, audioBytes, filename)
	if err != nil {
		if errors.Is(err, ai.ErrAudioTooLarge) {
			h.sendBestEffort(msg.PhoneNumber, "That voice message is too long for me to transcribe — try a shorter one or type it instead?")
		} else {
			slog.Error("whatsapp: transcription failed", "error", err, "phone", msg.PhoneNumber)
			h.sendBestEffort(msg.PhoneNumber, "Sorry, I couldn't understand that voice message — try typing instead?")
		}
		return "", false
	}
	if transcript == "" {
		h.sendBestEffort(msg.PhoneNumber, "I couldn't make out anything in that voice message — mind trying again, or typing it instead?")
		return "", false
	}
	return transcript, true
}

func extensionForMimetype(mimetypes ...string) string {
	for _, mt := range mimetypes {
		switch {
		case strings.Contains(mt, "ogg"):
			return "ogg"
		case strings.Contains(mt, "mp4"), strings.Contains(mt, "m4a"):
			return "m4a"
		case strings.Contains(mt, "mpeg"), strings.Contains(mt, "mp3"):
			return "mp3"
		case strings.Contains(mt, "wav"):
			return "wav"
		}
	}
	return "ogg" // WhatsApp voice notes are ogg/opus by default; safe fallback
}

func (h *WhatsAppHandler) handleUnlinkedMessage(ctx context.Context, msg inboundMessage) {
	if matches := linkCommandRe.FindStringSubmatch(msg.Text); matches != nil {
		code := matches[1]
		userID, err := h.whatsappRepo.ResolveLinkCode(ctx, code)
		if errors.Is(err, repository.ErrNotFound) {
			h.sendBestEffort(msg.PhoneNumber, "That code is invalid or expired. Generate a new one from Settings > WhatsApp in the app.")
			return
		}
		if err != nil {
			slog.Error("whatsapp: resolve link code failed", "error", err)
			h.sendBestEffort(msg.PhoneNumber, "Sorry, I hit an error — try again in a moment.")
			return
		}
		if _, err := h.whatsappRepo.CreateLink(ctx, userID, msg.PhoneNumber); err != nil {
			slog.Error("whatsapp: create link failed", "error", err)
			h.sendBestEffort(msg.PhoneNumber, "Sorry, I hit an error linking your account — try again in a moment.")
			return
		}
		h.sendBestEffort(msg.PhoneNumber, "Linked! You can now ask me about your habits and goals right here.")
		return
	}

	h.sendBestEffort(msg.PhoneNumber, "Hi! To use the Arete assistant, link your account first: open the app, go to Settings > WhatsApp, and send me the code shown there.")
}

func (h *WhatsAppHandler) sendBestEffort(phone, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := h.evoClient.SendText(ctx, phone, text); err != nil {
		slog.Error("whatsapp: send reply failed", "error", err, "phone", phone)
	}
}

func toOpenAIMessages(msgs []models.AIMessage) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs))
	for _, m := range msgs {
		var payload struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(m.Content, &payload); err != nil || payload.Text == "" {
			continue
		}
		if m.Role == "assistant" {
			out = append(out, openai.AssistantMessage(payload.Text))
		} else {
			out = append(out, openai.UserMessage(payload.Text))
		}
	}
	return out
}

// --- Linking endpoints (JWT-protected) ---

func (h *WhatsAppHandler) CreateLinkCode(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var linkCode models.WhatsAppLinkCode
	var err error
	for range 3 {
		var code string
		code, err = generateSixDigitCode()
		if err != nil {
			http.Error(w, "failed to generate code", http.StatusInternalServerError)
			return
		}
		linkCode, err = h.whatsappRepo.CreateLinkCode(r.Context(), authUser.ID, code, time.Now().Add(linkCodeTTL))
		if err == nil || !errors.Is(err, repository.ErrCodeTaken) {
			break
		}
	}
	if err != nil {
		http.Error(w, "failed to create link code", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"code":      linkCode.Code,
		"expiresAt": linkCode.ExpiresAt,
	})
}

func (h *WhatsAppHandler) Status(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	link, err := h.whatsappRepo.GetLinkByUser(r.Context(), authUser.ID)
	if errors.Is(err, repository.ErrNotFound) {
		writeJSON(w, http.StatusOK, map[string]any{"linked": false})
		return
	}
	if err != nil {
		http.Error(w, "failed to fetch status", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"linked":            true,
		"phoneNumberMasked": maskPhone(link.PhoneNumber),
		"linkedAt":          link.LinkedAt,
	})
}

func (h *WhatsAppHandler) Unlink(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if err := h.whatsappRepo.DeleteLink(r.Context(), authUser.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "no linked whatsapp number", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to unlink", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func generateSixDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "••••"
	}
	return "•••• " + phone[len(phone)-4:]
}
