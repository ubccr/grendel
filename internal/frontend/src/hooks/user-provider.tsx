import { useQueryClient } from "@tanstack/react-query";
import { useStoredValue } from "@/hooks/use-stored-value";
import { createContext, useContext, useEffect, useMemo } from "react";
import { toast } from "sonner";

const userStorageKey = "user";

export type User = {
  username: string;
  role: string;
  expire: number;
} | null;

type UserProviderProps = {
  children: React.ReactNode;
};

type UserProviderState = {
  user: User;
  setUser: (User: User) => void;
};

const initialState: UserProviderState = {
  user: null,
  setUser: () => null,
};

const UserProviderContext = createContext<UserProviderState>(initialState);

export function UserProvider({ children, ...props }: UserProviderProps) {
  // the user is stored, so signing in or out in one tab does the same in the others. Stored data that does not parse reads as signed out
  const [stored, setStored] = useStoredValue(userStorageKey);
  const user = useMemo<User>(() => {
    try {
      return stored ? (JSON.parse(stored) as User) : null;
    } catch {
      return null;
    }
  }, [stored]);
  const queryClient = useQueryClient();

  // a new session starts with an empty cache and ends when its token expires; the timer goes with the session, so signing out, here or in another tab, cancels it
  useEffect(() => {
    if (!user?.expire) return;
    queryClient.clear();
    const timer = setTimeout(() => {
      setStored(undefined);
      toast.warning("Session expired", {
        description: "Authentication token has expired, please login again",
      });
    }, user.expire - Date.now());
    return () => clearTimeout(timer);
  }, [user, queryClient, setStored]);

  const value = {
    user,
    setUser: (user: User) => setStored(user === null ? undefined : JSON.stringify(user)),
  };

  return (
    <UserProviderContext.Provider {...props} value={value}>
      {children}
    </UserProviderContext.Provider>
  );
}

export const useUser = () => {
  const context = useContext(UserProviderContext);

  if (context === undefined) throw new Error("useUser must be used within a UserProvider");

  return context;
};
