import paths from "./paths";

export class FetchClient {
  baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  async fetch<T>(path: string, options?: RequestInit): Promise<T> {
    try {
      const response = await fetch(`${this.baseURL}${path}`, {
        headers: {
          "Content-Type": "application/json",
        },
        ...options,
      });

      if (!response.ok) {
        const errorData = await response.json();
        if (errorData && errorData.message) {
          const error = new Error(errorData.message);
          (error as any).data = errorData;
          throw error;
        }

        throw new Error(
          `HTTP Error: ${response.status} ${response.statusText}`
        );
      }

      try {
        return await response.json();
      } catch (jsonError) {
        return {} as T;
      }
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error(`Fetch error: ${String(error)}`);
    }
  }
}
export const client = new FetchClient(`http://localhost:8080${paths.base}`);

