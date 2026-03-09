import { lazy, Suspense } from 'react'
import { Outlet, createRootRouteWithContext } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'
import type { AuthState } from '@/lib/auth'
import { Toaster } from '@/components/ui/sonner'
import { AchievementNotifications } from '@/components/achievement-notifications'


interface MyRouterContext {
  queryClient: QueryClient
  auth: AuthState
}

const DevTools = import.meta.env.DEV
  ? lazy(() =>
      Promise.all([
        import('@tanstack/react-router-devtools'),
        import('@tanstack/react-devtools'),
        import('../integrations/tanstack-query/devtools'),
      ]).then(([{ TanStackRouterDevtoolsPanel }, { TanStackDevtools }, { default: TanStackQueryDevtools }]) => ({
        default: function DevToolsPanel() {
          return (
            <TanStackDevtools
              config={{ position: 'bottom-right' }}
              plugins={[
                { name: 'Tanstack Router', render: <TanStackRouterDevtoolsPanel /> },
                TanStackQueryDevtools,
              ]}
            />
          )
        },
      }))
    )
  : () => null

export const Route = createRootRouteWithContext<MyRouterContext>()({
  component: () => (
    <>
      <Outlet />
      <Toaster />
      <AchievementNotifications />
      <Suspense>
        <DevTools />
      </Suspense>
    </>
  ),
})
