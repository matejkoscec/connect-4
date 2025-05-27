import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { FetchClient } from "./api/client";
import paths from "./api/paths";
import { Button } from "./components/ui/button";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

export const client = new FetchClient(`http://localhost:8080${paths.base}`);

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <div className="flex flex-col items-center justify-center min-h-svh">
        <Button>Click me</Button>
      </div>
    </QueryClientProvider>
  );
}
