import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import { api } from './api-client'

export interface AuthUser {
  userId: string
  username: string
}

export interface AuthState {
  user: AuthUser | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (user: AuthUser) => void
  logout: () => void
}

const AuthContext = createContext<AuthState | null>(null)
const SESSION_KEY = 'arete_user'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(() => {
    try {
      const stored = sessionStorage.getItem(SESSION_KEY)
      return stored ? JSON.parse(stored) : null
    } catch {
      return null
    }
  })
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    api.auth.me()
      .then((data) => {
        setUser(data)
        sessionStorage.setItem(SESSION_KEY, JSON.stringify(data))
      })
      .catch(() => {
        setUser(null)
        sessionStorage.removeItem(SESSION_KEY)
      })
      .finally(() => setIsLoading(false))
  }, [])

  useEffect(() => {
    function handleUnauthorized() {
      setUser(null)
      sessionStorage.removeItem(SESSION_KEY)
    }
    window.addEventListener('auth:unauthorized', handleUnauthorized)
    return () => window.removeEventListener('auth:unauthorized', handleUnauthorized)
  }, [])

  const login = useCallback((u: AuthUser) => {
    setUser(u)
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(u))
  }, [])

  const logout = useCallback(() => {
    setUser(null)
    sessionStorage.removeItem(SESSION_KEY)
  }, [])

  return (
    <AuthContext.Provider value={{
      user,
      isAuthenticated: user !== null,
      isLoading,
      login,
      logout,
    }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
