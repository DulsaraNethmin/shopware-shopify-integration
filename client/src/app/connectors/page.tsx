'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent } from '@/components/ui/Card';
import DashboardLayout from '@/components/layout/DashboardLayout';
import PageHeader from '@/components/layout/PageHeader';
import { AuthGuard } from '@/contexts/AuthContext';
import api, { Connector } from '@/lib/api';
import { formatDate } from '@/lib/utils';
import toast from 'react-hot-toast';

export default function ConnectorsPage() {
    const [connectors, setConnectors] = useState<Connector[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    const fetchConnectors = async () => {
        try {
            setIsLoading(true);
            const response = await api.getConnectors();
            setConnectors(response.data.data || []);
        } catch (error) {
            console.error('Error fetching connectors:', error);
            toast.error('Failed to load connectors');
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchConnectors();
    }, []);

    const handleDeleteConnector = async (id: number) => {
        if (!window.confirm('Are you sure you want to delete this connector?')) {
            return;
        }

        try {
            await api.deleteConnector(id);
            toast.success('Connector deleted successfully');
            fetchConnectors();
        } catch (error) {
            console.error('Error deleting connector:', error);
            toast.error('Failed to delete connector');
        }
    };

    const handleTestConnection = async (id: number) => {
        try {
            await api.testConnection(id);
            toast.success('Connection test successful');
        } catch (error) {
            console.error('Error testing connection:', error);
            toast.error('Connection test failed');
        }
    };

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title="Connectors"
                    description="Manage your Shopware and Shopify connections"
                    actions={
                        <Button asChild>
                            <Link href="/connectors/new">Add Connector</Link>
                        </Button>
                    }
                />

                {isLoading ? (
                    <div className="flex justify-center items-center h-64">
                        <p>Loading connectors...</p>
                    </div>
                ) : connectors.length === 0 ? (
                    <Card>
                        <CardContent className="flex flex-col items-center justify-center h-64">
                            <p className="text-muted-foreground mb-4">No connectors found</p>
                            <Button asChild>
                                <Link href="/connectors/new">Add Your First Connector</Link>
                            </Button>
                        </CardContent>
                    </Card>
                ) : (
                    <div className="grid gap-6">
                        {connectors.map((connector) => (
                            <Card key={connector.id} className="overflow-hidden">
                                <div className="flex flex-col md:flex-row md:items-center md:justify-between p-6">
                                    <div className="mb-4 md:mb-0">
                                        <div className="flex items-center gap-2">
                                            <h3 className="text-lg font-semibold">{connector.name}</h3>
                                            <Badge variant={connector.is_active ? 'success' : 'secondary'}>
                                                {connector.is_active ? 'Active' : 'Inactive'}
                                            </Badge>
                                            <Badge variant="info">{connector.type}</Badge>
                                        </div>
                                        <p className="text-sm text-muted-foreground">URL: {connector.url}</p>
                                        <p className="text-sm text-muted-foreground">Created: {formatDate(connector.created_at)}</p>
                                    </div>
                                    <div className="flex flex-wrap gap-2">
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={() => handleTestConnection(connector.id)}
                                        >
                                            Test Connection
                                        </Button>
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            asChild
                                        >
                                            <Link href={`/connectors/${connector.id}`}>Edit</Link>
                                        </Button>
                                        <Button
                                            variant="destructive"
                                            size="sm"
                                            onClick={() => handleDeleteConnector(connector.id)}
                                        >
                                            Delete
                                        </Button>
                                    </div>
                                </div>
                            </Card>
                        ))}
                    </div>
                )}
            </DashboardLayout>
        </AuthGuard>
    )
}