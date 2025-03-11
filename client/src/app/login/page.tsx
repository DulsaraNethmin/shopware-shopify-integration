'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { Toaster } from 'react-hot-toast';

export default function LoginPage() {
    const { isAuthenticated, isLoading, login } = useAuth();
    const router = useRouter();

    // Redirect to dashboard if already authenticated
    useEffect(() => {
        if (!isLoading && isAuthenticated) {
            router.push('/dashboard');
        }
    }, [isAuthenticated, isLoading, router]);

    const handleLogin = () => {
        login();
    };

    if (isLoading) {
        return (
            <div className="fixed inset-0 flex items-center justify-center bg-white">
                <div className="text-center">
                    <div className="animate-spin inline-block w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
                    <p className="mt-2 text-gray-600">Loading...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-100 p-4">
            <div className="w-full max-w-md bg-white rounded-xl shadow-md overflow-hidden">
                <div className="p-8 text-center">
                    <div className="mb-4 flex justify-center">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="48"
                            height="48"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="1.5"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            className="text-gray-800 w-12 h-12"
                        >
                            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                            <path d="M17 21h-6a2 2 0 0 1 0-4h4a2 2 0 0 0 0-4h-2V3"></path>
                        </svg>
                    </div>

                    <h1 className="text-2xl font-bold text-gray-900 mb-2">
                        Shopware to Shopify Integration
                    </h1>
                    <p className="text-gray-600 mb-6">
                        Secure access to your integration dashboard using your Keycloak credentials.
                    </p>

                    <button
                        onClick={handleLogin}
                        className="w-full py-3 bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors duration-300 cursor-pointer"
                    >
                        Sign in with Keycloak
                    </button>

                    <div className="mt-6 text-sm text-gray-500">
                        <p className="mb-1">Secure login powered by Keycloak</p>
                        <p>© {new Date().getFullYear()} Integration Platform</p>
                    </div>
                </div>
            </div>
            <Toaster position="top-right" />
        </div>
    );
}