import { AuthProvider, useAuth } from "@/contexts/AuthContext";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  createRootRoute,
  Link,
  Outlet,
  useNavigate,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { PowerCircle, User } from "lucide-react";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

function HeaderComponent() {
  const { user, logout, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  return (
    <header className="bg-gray-800 text-white p-4 flex justify-between items-center shadow-md">
      <Link
        to={isAuthenticated ? "/games" : "/login"}
        className="text-xl font-bold"
      >
        Connect 4
      </Link>
      <div className="flex items-center space-x-2">
        {user && (
          <>
            <User className="h-5 w-5" />
            <span className="text-lg font-medium">{user.username}</span>
            <PowerCircle
              onClick={() => {
                logout();
                navigate({ to: "/login" });
              }}
              className="ml-4 h-5 w-5 cursor-pointer"
            />
          </>
        )}
      </div>
    </header>
  );
}

export const Route = createRootRoute({
  component: () => (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <main className="flex flex-col relative min-h-screen">
          <HeaderComponent />
          <Outlet />
          <TanStackRouterDevtools />
        </main>
      </AuthProvider>
    </QueryClientProvider>
  ),
});
