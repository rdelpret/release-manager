import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import * as api from "@/lib/api";
import type { Campaign, TemplateType, User } from "@/lib/types";

export function useCampaigns() {
  return useQuery<Campaign[]>({
    queryKey: ["campaigns"],
    queryFn: api.listCampaigns,
    staleTime: 30_000,
  });
}

export function useCampaign(id: string) {
  return useQuery<Campaign>({
    queryKey: ["campaign", id],
    queryFn: () => api.getCampaign(id),
    enabled: !!id && id !== "_",
  });
}

export function useCreateCampaign() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ name, releaseDate, templateType }: { name: string; releaseDate?: string; templateType?: TemplateType }) =>
      api.createCampaign(name, releaseDate, templateType),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["campaigns"] }),
  });
}

export function useDuplicateCampaign() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.duplicateCampaign(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["campaigns"] }),
  });
}

export function useArchiveCampaign() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, archived }: { id: string; archived: boolean }) =>
      api.archiveCampaign(id, archived),
    onMutate: async ({ id, archived }) => {
      await queryClient.cancelQueries({ queryKey: ["campaigns"] });
      const prev = queryClient.getQueryData<Campaign[]>(["campaigns"]);
      queryClient.setQueryData<Campaign[]>(["campaigns"], (old) =>
        old?.map((c) => (c.id === id ? { ...c, archived } : c))
      );
      return { prev };
    },
    onError: (_err, _vars, context) => {
      if (context?.prev) queryClient.setQueryData(["campaigns"], context.prev);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: ["campaigns"] }),
  });
}

export function useDeleteCampaign() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deleteCampaign(id),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ["campaigns"] });
      const prev = queryClient.getQueryData<Campaign[]>(["campaigns"]);
      queryClient.setQueryData<Campaign[]>(["campaigns"], (old) =>
        old?.filter((c) => c.id !== id)
      );
      return { prev };
    },
    onError: (_err, _id, context) => {
      if (context?.prev) queryClient.setQueryData(["campaigns"], context.prev);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: ["campaigns"] }),
  });
}

export function useUsers() {
  return useQuery<User[]>({
    queryKey: ["users"],
    queryFn: api.listUsers,
    staleTime: 10 * 60 * 1000, // 10 minutes — users rarely change
  });
}
