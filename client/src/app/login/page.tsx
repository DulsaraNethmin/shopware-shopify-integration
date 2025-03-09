'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { Button } from '@/components/ui/Button';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/Card';
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
            <div className="flex items-center justify-center min-h-screen">
                <p>Loading...</p>
            </div>
        );
    }

    return (
        <div className="flex items-center justify-center min-h-screen bg-gray-100">
            <Card className="w-full max-w-md">
                <CardHeader className="space-y-1">
                    <CardTitle className="text-2xl font-bold text-center">
                        Shopware to Shopify Integration
                    </CardTitle>
                    <CardDescription className="text-center">
                        Sign in to access the dashboard
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="text-center">
                        <p className="text-sm text-muted-foreground mb-6">
                            Click the button below to log in using your Keycloak account.
                        </p>
                    </div>
                </CardContent>
                <CardFooter>
                    <Button
                        onClick={handleLogin}
                        className="w-full"
                    >
                        Sign in with Keycloak
                    </Button>
                </CardFooter>
            </Card>
            <Toaster position="top-right" />
        </div>
    );
}