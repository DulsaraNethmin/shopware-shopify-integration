'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { Button } from '@/components/ui/Button';

export default function Navbar() {
    const pathname = usePathname();
    const { isAuthenticated, logout, userInfo } = useAuth();

    if (!isAuthenticated) return null;

    return (
        <nav className="bg-white border-b border-gray-200 fixed z-30 w-full">
            <div className="px-3 py-3 lg:px-5 lg:pl-3">
                <div className="flex items-center justify-between">
                    <div className="flex items-center justify-start">
                        <Link href="/dashboard" className="flex ml-2 md:mr-24">
              <span className="self-center text-xl font-semibold sm:text-2xl whitespace-nowrap">
                Shopware ↔ Shopify Integration
              </span>
                        </Link>
                    </div>
                    <div className="flex items-center">
                        <Button
                            onClick={logout}
                            variant="ghost"
                            className="ml-3 cursor-pointer"
                        >
                            <span className="mr-2">{userInfo?.email || userInfo?.name || userInfo?.username}</span>
                            Logout
                        </Button>
                    </div>
                </div>
            </div>
        </nav>
    );
}