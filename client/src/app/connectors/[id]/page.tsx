'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card';
import DashboardLayout from '@/components/layout/DashboardLayout';
import PageHeader from '@/components/layout/PageHeader';
import { AuthGuard } from '@/contexts/AuthContext';
import api, { Connector } from '@/lib/api';
import { formatError } from '@/lib/utils';
import toast from 'react-hot-toast';

export default function EditConnectorPage() {
    const params = useParams();
    const connectorId = parseInt(params.id as string);
    const router = useRouter();

    const [isLoading, setIsLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const [connector, setConnector] = useState<Partial<Connector>>({
        name: '',
        type: 'shopware',
        url: '',
        username: '',
        is_active: true
    });
    const [apiKey, setApiKey] = useState('');
    const [apiSecret, setApiSecret] = useState('');
    const [accessToken, setAccessToken] = useState('');
    const [password, setPassword] = useState('');

    useEffect(() => {
        async function fetchConnector() {
            try {
                const response = await api.getConnector(connectorId);
                const connectorData = response.data.data;
                setConnector({
                    id: connectorData.id,
                    name: connectorData.name,
                    type: connectorData.type,
                    url: connectorData.url,
                    username: connectorData.username,
                    is_active: connectorData.is_active
                });
            } catch (error) {
                console.error('Error fetching connector:', error);
                toast.error('Failed to load connector details');
                router.push('/connectors');
            } finally {
                setIsLoading(false);
            }
        }

        fetchConnector();
    }, [connectorId, router]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        const { name, value, type } = e.target;

        if (type === 'checkbox') {
            const checked = (e.target as HTMLInputElement).checked;
            setConnector({ ...connector, [name]: checked });
        } else {
            setConnector({ ...connector, [name]: value });
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSaving(true);

        try {
            // Only include credentials if they've been modified
            const updateData: any = { ...connector };

            if (apiKey) updateData.api_key = apiKey;
            if (apiSecret) updateData.api_secret = apiSecret;
            if (accessToken) updateData.access_token = accessToken;
            if (password) updateData.password = password;

            await api.updateConnector(connectorId, updateData);
            toast.success('Connector updated successfully');
            router.push('/connectors');
        } catch (error) {
            console.error('Error updating connector:', error);
            toast.error(`Failed to update connector: ${formatError(error)}`);
        } finally {
            setIsSaving(false);
        }
    };

    const handleTestConnection = async () => {
        try {
            await api.testConnection(connectorId);
            toast.success('Connection test successful');
        } catch (error) {
            console.error('Error testing connection:', error);
            toast.error(`Connection test failed: ${formatError(error)}`);
        }
    };

    if (isLoading) {
        return (
            <AuthGuard>
                <DashboardLayout>
                    <div className="flex justify-center items-center h-64">
                        <p>Loading connector details...</p>
                    </div>
                </DashboardLayout>
            </AuthGuard>
        );
    }

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title="Edit Connector"
                    description="Update your connection details"
                    actions={
                        <Button
                            variant="outline"
                            onClick={handleTestConnection}
                        >
                            Test Connection
                        </Button>
                    }
                />

                <Card className="max-w-2xl mx-auto">
                    <form onSubmit={handleSubmit}>
                        <CardHeader>
                            <CardTitle>Edit Connector</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <label htmlFor="name" className="text-sm font-medium">
                                    Name
                                </label>
                                <Input
                                    id="name"
                                    name="name"
                                    value={connector.name}
                                    onChange={handleChange}
                                    required
                                />
                            </div>

                            <div className="space-y-2">
                                <label htmlFor="type" className="text-sm font-medium">
                                    Type
                                </label>
                                <Select
                                    id="type"
                                    name="type"
                                    value={connector.type}
                                    onChange={handleChange}
                                    disabled // Cannot change connector type after creation
                                >
                                    <option value="shopware">Shopware</option>
                                    <option value="shopify">Shopify</option>
                                </Select>
                            </div>

                            <div className="space-y-2">
                                <label htmlFor="url" className="text-sm font-medium">
                                    URL
                                </label>
                                <Input
                                    id="url"
                                    name="url"
                                    value={connector.url}
                                    onChange={handleChange}
                                    required
                                />
                            </div>

                            {connector.type === 'shopware' && (
                                <>
                                    <div className="space-y-2">
                                        <label htmlFor="api_key" className="text-sm font-medium">
                                            API Key (leave blank to keep current)
                                        </label>
                                        <Input
                                            id="api_key"
                                            value={apiKey}
                                            onChange={(e) => setApiKey(e.target.value)}
                                            type="password"
                                            placeholder="••••••••"
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <label htmlFor="api_secret" className="text-sm font-medium">
                                            API Secret (leave blank to keep current)
                                        </label>
                                        <Input
                                            id="api_secret"
                                            value={apiSecret}
                                            onChange={(e) => setApiSecret(e.target.value)}
                                            type="password"
                                            placeholder="••••••••"
                                        />
                                    </div>
                                </>
                            )}

                            {connector.type === 'shopify' && (
                                <>
                                    <div className="space-y-2">
                                        <label htmlFor="access_token" className="text-sm font-medium">
                                            Access Token (leave blank to keep current)
                                        </label>
                                        <Input
                                            id="access_token"
                                            value={accessToken}
                                            onChange={(e) => setAccessToken(e.target.value)}
                                            type="password"
                                            placeholder="••••••••"
                                        />
                                    </div>
                                </>
                            )}

                            <div className="space-y-2">
                                <label htmlFor="username" className="text-sm font-medium">
                                    Username (Optional)
                                </label>
                                <Input
                                    id="username"
                                    name="username"
                                    value={connector.username}
                                    onChange={handleChange}
                                />
                            </div>

                            <div className="flex items-center">
                                <input
                                    id="is_active"
                                    name="is_active"
                                    type="checkbox"
                                    checked={connector.is_active}
                                    onChange={(e) => setConnector({...connector, is_active: e.target.checked})}
                                    className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                />
                                <label htmlFor="is_active" className="ml-2 block text-sm text-gray-900">
                                    Active
                                </label>
                            </div>
                        </CardContent>
                        <CardFooter className="flex justify-between">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => router.push('/connectors')}
                            >
                                Cancel
                            </Button>
                            <Button type="submit" disabled={isSaving}>
                                {isSaving ? 'Saving...' : 'Save Changes'}
                            </Button>
                        </CardFooter>
                    </form>
                </Card>
            </DashboardLayout>
        </AuthGuard>
    );
}