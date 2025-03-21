import { useQueryClient } from "@tanstack/react-query";
import { createContext, useContext, useEffect, useState } from "react";
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
  const [id, setId] = useState<NodeJS.Timeout | undefined>(undefined);
  const [user, setUser] = useState<User>(() => {
    const user = localStorage.getItem(userStorageKey);
    if (user == null) return null;
    return JSON.parse(user) as User;
  });
  const queryClient = useQueryClient();

  useEffect(() => {
    if (user != null && user?.expire) {
      const difference = new Date(user.expire).getTime() - new Date().getTime();
      queryClient.clear();
      setId(
        setTimeout(() => {
          setUser(null);
          localStorage.removeItem(userStorageKey);
          toast.warning("Session expired", {
            description: "Authentication token has expired, please login again",
          });
        }, difference)
      );
    }
  }, [user]);

  const value = {
    user,
    setUser: (user: User) => {
      if (user !== null) {
        localStorage.setItem(userStorageKey, JSON.stringify(user));
      } else {
        clearTimeout(id);
        localStorage.removeItem(userStorageKey);
      }

      setUser(user);
    },
  };

  return (
    <UserProviderContext.Provider {...props} value={value}>
      {children}
    </UserProviderContext.Provider>
  );
}

export const useUser = () => {
  const context = useContext(UserProviderContext);

  if (context === undefined)
    throw new Error("useUser must be used within a UserProvider");

  return context;
};
