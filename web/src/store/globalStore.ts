import { create } from 'zustand';
import { persist, devtools } from 'zustand/middleware';
import type { User } from '../types/user';

interface GlobalStore {
  user: User | null;
  mapID: string | null;
  loading: boolean;
  error: string | null;
  setUser: (user: User | null) => void;
  setMapID: (id: string | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useGlobalStore = create<GlobalStore>()(
  devtools(
    persist(
      (set) => ({
        user: null,
        mapID: null,
        loading: false,
        error: null,
        setUser: (user) => set({ user }),
        setMapID: (id) => set({ mapID: id }),
        setLoading: (loading) => set({ loading }),
        setError: (error) => set({ error }),
      }),
      { name: 'global-store' }
    )
  )
); 