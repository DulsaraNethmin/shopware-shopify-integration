'use client';

import React, { createContext, useContext, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
    getKeycloakInstance,
    initKeycloak,
    isAuthenticated,
    login as keycloakLogin,
    logout as keycloakLogout,
    getUserInfo
} from '@/lib/keycloak';
import toast from 'react-hot-toast';

interface UserInfo {
    username?: string;
    email?: string;
    name?: string;
    roles?: string[];
}

interface AuthContextType {
    isAuthenticated: boolean;
    isLoading: boolean;
    userInfo: UserInfo | null;
    login: () => void;
    logout: () => void;
}

const AuthContext = createContext<AuthContextType>({
    isAuthenticated: false,
    isLoading: true,
    userInfo: null,
    login: () => {},
    logout: () => {},
});

export const useAuth = () => useContext(AuthContext);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [isLoading, setIsLoading] = useState(true);
    const [authState, setAuthState] = useState({
        isAuthenticated: false,
        userInfo: null as UserInfo | null
    });
    const router = useRouter();

    // Initialize Keycloak on mount
    useEffect(() => {
        const initAuth = async () => {
            try {
                // Only initialize if window is defined (client-side)
                if (typeof window !== 'undefined') {
                    const authenticated = await initKeycloak();

                    if (authenticated) {
                        const userInfo = getUserInfo();
                        setAuthState({
                            isAuthenticated: true,
                            userInfo
                        });
                    } else {
                        setAuthState({
                            isAuthenticated: false,
                            userInfo: null
                        });
                    }
                }
            } catch (error) {
                console.error('Error initializing Keycloak:', error);
                setAuthState({
                    isAuthenticated: false,
                    userInfo: null
                });
            } finally {
                setIsLoading(false);
            }
        };

        initAuth();

        // Set up event listener for Keycloak auth state changes
        const keycloak = getKeycloakInstance();
        if (keycloak) {
            keycloak.onAuthSuccess = () => {
                setAuthState({
                    isAuthenticated: true,
                    userInfo: getUserInfo()
                });
            };

            keycloak.onAuthError = () => {
                setAuthState({
                    isAuthenticated: false,
                    userInfo: null
                });
            };

            keycloak.onAuthRefreshSuccess = () => {
                setAuthState({
                    isAuthenticated: true,
                    userInfo: getUserInfo()
                });
            };

            keycloak.onAuthLogout = () => {
                setAuthState({
                    isAuthenticated: false,
                    userInfo: null
                });
                router.push('/login');
            };
        }
    }, [router]);

    const login = () => {
        keycloakLogin();
    };

    const logout = () => {
        keycloakLogout();
        toast.success('Logged out successfully');
    };

    return (
        <AuthContext.Provider
            value={{
                isAuthenticated: authState.isAuthenticated,
                isLoading,
                userInfo: authState.userInfo,
                login,
                logout
            }}
        >
            {children}
        </AuthContext.Provider>
    );
};

// Auth guard component to protect routes
export const AuthGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { isAuthenticated, isLoading } = useAuth();
    const router = useRouter();

    useEffect(() => {
        if (!isLoading && !isAuthenticated) {
            router.push('/login');
        }
    }, [isAuthenticated, isLoading, router]);

    if (isLoading) {
        return <div className="flex items-center justify-center min-h-screen">Loading...</div>;
    }

    if (!isAuthenticated) {
        return null;
    }

    return <>{children}</>;
};