import { useState, useEffect } from "react";
import React from "react";
import { useLocation } from "react-router";

interface IUser {
  id: string;
  email: string;
  name: string;
  picture: string;
}

interface IUserState {
  user: IUser | null;
  isAuthenticated: boolean;
  accessToken: string | undefined;
}

interface IAuthContext {
  userState: IUserState | null;
  login: () => void;
}

export const AuthContext = React.createContext<IAuthContext | undefined>(undefined);
const authServerUrl = "http://localhost:4000";

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [userState, setUserState] = useState<IUserState | null>(null);
  const location = useLocation();

  useEffect(() => {
    const queryParams = new URLSearchParams(location.search);
    const accessToken = queryParams.get('access_token');

    if (accessToken) {
      const fetchUser = async () => {
        try {
          const res = await fetch(`${authServerUrl}/auth/userinfo`, {
            credentials: "include",
            headers: {
              Authorization: `Bearer ${accessToken}`,
            },
          });
          const data: IUser = await res.json();
          setUserState({ user: data, isAuthenticated: true, accessToken: accessToken });
        } catch (error) {
          console.error(error);
        }
      };

      fetchUser();
    }
  }, [location]);

  const login = async () => {
    window.location.href = `${authServerUrl}/auth/google`;
  };


  return (
    <AuthContext.Provider value={{ userState, login }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => {
  const context = React.useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}