"use client"

import { useState } from "react"
import { Plus, Pencil, Trash2 } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { PomodoroProjectDialog } from "@/components/pomodoro-project-dialog"
import {
  useProjects,
  useCreateProject,
  useUpdateProject,
  useDeleteProject,
} from "@/hooks/use-pomodoro"
import type { PomodoroProject } from "@/lib/types"

export function PomodoroProjectsPanel() {
  const { data: projects = [] } = useProjects()
  const createProject = useCreateProject()
  const updateProject = useUpdateProject()
  const deleteProject = useDeleteProject()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingProject, setEditingProject] = useState<PomodoroProject | null>(null)

  function handleOpenCreate() {
    setEditingProject(null)
    setDialogOpen(true)
  }

  function handleOpenEdit(project: PomodoroProject) {
    setEditingProject(project)
    setDialogOpen(true)
  }

  function handleSave(name: string, color: string) {
    if (editingProject) {
      updateProject.mutate({ id: editingProject.id, name, color })
    } else {
      createProject.mutate({ name, color })
    }
  }

  return (
    <Card className="border-foreground/20 h-full">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="font-display text-lg font-semibold tracking-tight">
          Projects
        </CardTitle>
        <Button
          variant="outline"
          size="sm"
          onClick={handleOpenCreate}
          className="h-8 gap-1.5 border-border"
        >
          <Plus className="w-4 h-4" />
          New
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-1">
        {projects.length === 0 ? (
          <p className="text-sm text-muted-foreground py-6 text-center">
            No projects yet — create one to start tracking time against it.
          </p>
        ) : (
          projects.map((project) => (
            <div
              key={project.id}
              className="group flex items-center gap-2.5 px-2 py-2 rounded-lg hover:bg-muted"
            >
              <span
                className="w-2.5 h-2.5 rounded-full shrink-0"
                style={{ backgroundColor: project.color }}
              />
              <span className="text-sm flex-1 truncate">{project.name}</span>
              <button
                onClick={() => handleOpenEdit(project)}
                className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-foreground transition-opacity"
                aria-label={`Edit ${project.name}`}
              >
                <Pencil className="w-3.5 h-3.5" />
              </button>
              <button
                onClick={() => deleteProject.mutate(project.id)}
                className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity"
                aria-label={`Delete ${project.name}`}
              >
                <Trash2 className="w-3.5 h-3.5" />
              </button>
            </div>
          ))
        )}
      </CardContent>

      <PomodoroProjectDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        project={editingProject}
        onSave={handleSave}
      />
    </Card>
  )
}
