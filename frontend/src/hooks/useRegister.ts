import paths from "@/api/paths";
import { client } from "@/routes/__root";
import { useMutation } from "@tanstack/react-query";

export type RegisterUserRequest = {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
};

export type RegisterUserResponse = {
  message?: string;
};

export const useRegister = () => {
  return useMutation({
    mutationFn: async (
      userData: Omit<RegisterUserRequest, "confirmPassword">
    ) => {
      return client.fetch<RegisterUserResponse>(paths.auth.register, {
        method: "POST",
        body: JSON.stringify(userData),
      });
    },
  });
};
