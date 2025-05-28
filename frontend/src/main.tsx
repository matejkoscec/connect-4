import "@/index.css";
import {
  RouterProvider,
  createRouter,
  useNavigate,
} from "@tanstack/react-router";
import { StrictMode, useEffect } from "react";
import ReactDOM from "react-dom/client";

import { routeTree } from "./routeTree.gen";

const router = createRouter({
  routeTree,
  defaultNotFoundComponent: () => {
    const navigate = useNavigate();
    useEffect(() => {
      navigate({ to: "/login", replace: true });
    }, [navigate]);

    return <></>;
  },
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

const rootElement = document.getElementById("root")!;
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <StrictMode>
      <RouterProvider router={router} />
    </StrictMode>
  );
}
