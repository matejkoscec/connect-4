import { jwtDecode } from "jwt-decode";
import {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useState,
} from "react";

export interface DecodedToken {
  userId: string;
  username: string;
  iat: number;
  exp: number;
}

interface AuthContextType {
  token: string | null;
  user: DecodedToken | null;
  isAuthenticated: boolean;
  login: (token: string) => void;
  logout: () => void;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<DecodedToken | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const storedToken = localStorage.getItem("auth_token");
    if (storedToken) {
      try {
        const decoded = jwtDecode<DecodedToken>(storedToken);

        if (decoded.exp * 1000 > Date.now()) {
          setToken(storedToken);
          setUser(decoded);
        } else {
          localStorage.removeItem("auth_token");
        }
      } catch (error) {
        console.error("Failed to decode token on initial load:", error);
        localStorage.removeItem("auth_token");
      }
    }
    setIsLoading(false);
  }, []);

  const login = (newToken: string) => {
    try {
      const decoded = jwtDecode<DecodedToken>(newToken);
      setToken(newToken);
      setUser(decoded);
      localStorage.setItem("auth_token", newToken);
    } catch (error) {
      console.error("Login failed: could not decode new token.", error);
    }
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    localStorage.removeItem("auth_token");
  };

  const isAuthenticated = !isLoading && !!token && !!user;

  return (
    <AuthContext.Provider
      value={{
        token,
        user,
        isAuthenticated,
        login,
        logout,
        isLoading,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
