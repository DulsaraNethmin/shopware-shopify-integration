'use client';

import { useAuth } from '@/contexts/AuthContext';
import Navbar from './Navbar';
import Sidebar from './Sidebar';
import { Toaster } from 'react-hot-toast';

interface DashboardLayoutProps {
    children: React.ReactNode;
}

export default function DashboardLayout({ children }: DashboardLayoutProps) {
    const { isAuthenticated } = useAuth();

    if (!isAuthenticated) {
        return null;
    }

    return (
        <div className="flex overflow-hidden bg-gray-50">
            <Navbar />
            <Sidebar />
            <div className="h-full w-full bg-gray-50 relative overflow-y-auto lg:ml-64 pt-16">
                <main className="p-6">
                    {children}
                </main>
            </div>
            <Toaster position="top-right" />
        </div>
    );
}