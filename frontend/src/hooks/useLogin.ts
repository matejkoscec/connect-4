import paths from "@/api/paths";
import { useAuth } from "@/contexts/AuthContext";
import { client } from "@/routes/__root";
import { useMutation } from "@tanstack/react-query";

export type LoginUserRequest = {
  username: string;
  password: string;
};

export type LoginUserResponse = {
  token: string;
};

export const useLogin = () => {
  const { login } = useAuth();

  return useMutation({
    mutationFn: async (credentials: LoginUserRequest) => {
      return client.fetch<LoginUserResponse>(paths.auth.login, {
        method: "POST",
        body: JSON.stringify(credentials),
      });
    },
    onSuccess: (data) => {
      login(data.token);
    },
  });
};
