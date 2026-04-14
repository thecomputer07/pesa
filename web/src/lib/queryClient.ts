import { QueryClient } from "@tanstack/react-query";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      gcTime: 30 * 60 * 1000,
      retry: (failureCount, error: unknown) => {
        // Don't retry 4xx
        if (error instanceof Error && /4\d{2}/.test(error.message)) return false;
        return failureCount < 2;
      },
      refetchOnWindowFocus: false,
    },
  },
});
