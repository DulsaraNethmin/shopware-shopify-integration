'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
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

export default function NewConnectorPage() {
    const [isLoading, setIsLoading] = useState(false);
    const [formData, setFormData] = useState<Partial<Connector>>({
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

    const router = useRouter();

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        const { name, value, type } = e.target;

        if (type === 'checkbox') {
            const checked = (e.target as HTMLInputElement).checked;
            setFormData({ ...formData, [name]: checked });
        } else {
            setFormData({ ...formData, [name]: value });
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);

        try {
            // Prepare the full data
            const connectorData = {
                ...formData,
                api_key: apiKey,
                api_secret: apiSecret,
                access_token: accessToken,
                password: password
            };

            await api.createConnector(connectorData);
            toast.success('Connector created successfully');
            router.push('/connectors');
        } catch (error) {
            console.error('Error creating connector:', error);
            toast.error(`Failed to create connector: ${formatError(error)}`);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title="Add Connector"
                    description="Connect to Shopware or Shopify"
                />

                <Card className="max-w-2xl mx-auto">
                    <form onSubmit={handleSubmit}>
                        <CardHeader>
                            <CardTitle>New Connector</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <label htmlFor="name" className="text-sm font-medium">
                                    Name
                                </label>
                                <Input
                                    id="name"
                                    name="name"
                                    value={formData.name}
                                    onChange={handleChange}
                                    placeholder="My Shopware Store"
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
                                    value={formData.type}
                                    onChange={handleChange}
                                    required
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
                                    value={formData.url}
                                    onChange={handleChange}
                                    placeholder={formData.type === 'shopware' ? 'https://my-shopware-store.com' : 'my-shopify-store.myshopify.com'}
                                    required
                                />
                            </div>

                            {formData.type === 'shopware' && (
                                <>
                                    <div className="space-y-2">
                                        <label htmlFor="api_key" className="text-sm font-medium">
                                            API Key
                                        </label>
                                        <Input
                                            id="api_key"
                                            value={apiKey}
                                            onChange={(e) => setApiKey(e.target.value)}
                                            type="password"
                                            placeholder="Shopware API Key"
                                            required
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <label htmlFor="api_secret" className="text-sm font-medium">
                                            API Secret
                                        </label>
                                        <Input
                                            id="api_secret"
                                            value={apiSecret}
                                            onChange={(e) => setApiSecret(e.target.value)}
                                            type="password"
                                            placeholder="Shopware API Secret"
                                            required
                                        />
                                    </div>
                                </>
                            )}

                            {formData.type === 'shopify' && (
                                <>
                                    <div className="space-y-2">
                                        <label htmlFor="access_token" className="text-sm font-medium">
                                            Access Token
                                        </label>
                                        <Input
                                            id="access_token"
                                            value={accessToken}
                                            onChange={(e) => setAccessToken(e.target.value)}
                                            type="password"
                                            placeholder="Shopify Access Token"
                                            required
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
                                    value={formData.username}
                                    onChange={handleChange}
                                    placeholder="Username for reference"
                                />
                            </div>

                            <div className="flex items-center">
                                <input
                                    id="is_active"
                                    name="is_active"
                                    type="checkbox"
                                    checked={formData.is_active}
                                    onChange={(e) => setFormData({...formData, is_active: e.target.checked})}
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
                            <Button type="submit" disabled={isLoading}>
                                {isLoading ? 'Creating...' : 'Create Connector'}
                            </Button>
                        </CardFooter>
                    </form>
                </Card>
            </DashboardLayout>
        </AuthGuard>
    );
}