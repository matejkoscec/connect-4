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
        throw new Error(
          `HTTP Error: ${response.status} ${response.statusText}`
        );
      }

      try {
        return await response.json();
      } catch (jsonError) {
        throw new Error("Failed to parse JSON response");
      }
    } catch (error) {
      throw new Error(`Fetch error: ${(error as Error).message}`);
    }
  }
}
