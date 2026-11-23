import { useStoredValue } from "@/hooks/use-stored-value";
import { createContext, useCallback, useContext, useEffect, useMemo } from "react";
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
  signIn: (user: NonNullable<User>) => void;
  // signOut ends the session in this browser. Whoever calls it has already ended it on the server, or the server has by expiring the token
  signOut: () => void;
};

const initialState: UserProviderState = {
  user: null,
  signIn: () => null,
  signOut: () => null,
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
  const signOut = useCallback(() => setStored(undefined), [setStored]);

  // the session ends when its token expires; the timer goes with the session, so signing out, here or in another tab, cancels it
  useEffect(() => {
    if (!user?.expire) return;
    const timer = setTimeout(() => {
      signOut();
      toast.warning("Session expired", {
        description: "Authentication token has expired, please login again",
      });
    }, user.expire - Date.now());
    return () => clearTimeout(timer);
  }, [user, signOut]);

  const value = {
    user,
    signIn: (user: NonNullable<User>) => setStored(JSON.stringify(user)),
    signOut,
  };

  return (
    <UserProviderContext.Provider {...props} value={value}>
      {children}
    </UserProviderContext.Provider>
  );
}

export const useUser = () => useContext(UserProviderContext);
