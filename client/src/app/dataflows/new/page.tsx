'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/Card';
import DashboardLayout from '@/components/layout/DashboardLayout';
import PageHeader from '@/components/layout/PageHeader';
import { AuthGuard } from '@/contexts/AuthContext';
import api, { Connector, Dataflow } from '@/lib/api';
import { formatError } from '@/lib/utils';
import toast from 'react-hot-toast';

export default function NewDataflowPage() {
    const [isLoading, setIsLoading] = useState(true);
    const [isSaving, setIsSaving] = useState(false);
    const [shopwareConnectors, setShopwareConnectors] = useState<Connector[]>([]);
    const [shopifyConnectors, setShopifyConnectors] = useState<Connector[]>([]);

    const [formData, setFormData] = useState<Partial<Dataflow>>({
        name: '',
        description: '',
        type: 'product',
        status: 'active',
        source_connector_id: 0,
        dest_connector_id: 0
    });

    const router = useRouter();

    useEffect(() => {
        async function fetchConnectors() {
            try {
                const response = await api.getConnectors();
                const connectors = response.data.data || [];

                // Filter active connectors by type
                const shopware = connectors.filter(c => c.type === 'shopware' && c.is_active);
                const shopify = connectors.filter(c => c.type === 'shopify' && c.is_active);

                setShopwareConnectors(shopware);
                setShopifyConnectors(shopify);

                // Set default connectors if available
                if (shopware.length > 0) {
                    setFormData(prev => ({ ...prev, source_connector_id: shopware[0].id }));
                }

                if (shopify.length > 0) {
                    setFormData(prev => ({ ...prev, dest_connector_id: shopify[0].id }));
                }
            } catch (error) {
                console.error('Error fetching connectors:', error);
                toast.error('Failed to load connectors');
            } finally {
                setIsLoading(false);
            }
        }

        fetchConnectors();
    }, []);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        const { name, value } = e.target;

        // Handle numeric fields
        if (name === 'source_connector_id' || name === 'dest_connector_id') {
            setFormData({ ...formData, [name]: parseInt(value) });
        } else {
            setFormData({ ...formData, [name]: value });
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSaving(true);

        try {
            if (!formData.source_connector_id || !formData.dest_connector_id) {
                throw new Error('Please select both source and destination connectors');
            }

            await api.createDataflow(formData);
            toast.success('Dataflow created successfully');
            router.push('/dataflows');
        } catch (error) {
            console.error('Error creating dataflow:', error);
            toast.error(`Failed to create dataflow: ${formatError(error)}`);
        } finally {
            setIsSaving(false);
        }
    };

    const noConnectorsAvailable = shopwareConnectors.length === 0 || shopifyConnectors.length === 0;

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title="Create Dataflow"
                    description="Set up a new data migration flow"
                />

                {isLoading ? (
                    <div className="flex justify-center items-center h-64">
                        <p>Loading connector data...</p>
                    </div>
                ) : noConnectorsAvailable ? (
                    <Card>
                        <CardContent className="py-8">
                            <div className="text-center">
                                <h3 className="text-lg font-medium mb-2">Cannot Create Dataflow</h3>
                                <p className="text-muted-foreground mb-4">
                                    You need at least one active Shopware connector and one active Shopify connector to create a dataflow.
                                </p>
                                <Button asChild className="text-gray-700">
                                    <a href="/connectors/new">Create Connector</a>
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                ) : (
                    <Card className="max-w-2xl mx-auto">
                        <form onSubmit={handleSubmit}>
                            <CardHeader>
                                <CardTitle>New Dataflow</CardTitle>
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
                                        placeholder="Product Migration"
                                        required
                                    />
                                </div>

                                <div className="space-y-2">
                                    <label htmlFor="description" className="text-sm font-medium">
                                        Description
                                    </label>
                                    <textarea
                                        id="description"
                                        name="description"
                                        value={formData.description}
                                        onChange={handleChange}
                                        placeholder="Brief description of this dataflow"
                                        className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                        rows={3}
                                    />
                                </div>

                                <div className="space-y-2">
                                    <label htmlFor="type" className="text-sm font-medium">
                                        Data Type
                                    </label>
                                    <Select
                                        id="type"
                                        name="type"
                                        value={formData.type}
                                        onChange={handleChange}
                                        required
                                    >
                                        <option value="product">Product</option>
                                        <option value="order">Order</option>
                                    </Select>
                                </div>

                                <div className="space-y-2">
                                    <label htmlFor="source_connector_id" className="text-sm font-medium">
                                        Source Connector (Shopware)
                                    </label>
                                    <Select
                                        id="source_connector_id"
                                        name="source_connector_id"
                                        value={formData.source_connector_id?.toString()}
                                        onChange={handleChange}
                                        required
                                    >
                                        {shopwareConnectors.map(connector => (
                                            <option key={connector.id} value={connector.id}>
                                                {connector.name} - {connector.url}
                                            </option>
                                        ))}
                                    </Select>
                                </div>

                                <div className="space-y-2">
                                    <label htmlFor="dest_connector_id" className="text-sm font-medium">
                                        Destination Connector (Shopify)
                                    </label>
                                    <Select
                                        id="dest_connector_id"
                                        name="dest_connector_id"
                                        value={formData.dest_connector_id?.toString()}
                                        onChange={handleChange}
                                        required
                                    >
                                        {shopifyConnectors.map(connector => (
                                            <option key={connector.id} value={connector.id}>
                                                {connector.name} - {connector.url}
                                            </option>
                                        ))}
                                    </Select>
                                </div>

                                <div className="space-y-2">
                                    <label htmlFor="status" className="text-sm font-medium">
                                        Status
                                    </label>
                                    <Select
                                        id="status"
                                        name="status"
                                        value={formData.status}
                                        onChange={handleChange}
                                        required
                                    >
                                        <option value="active">Active</option>
                                        <option value="inactive">Inactive</option>
                                    </Select>
                                </div>
                            </CardContent>
                            <CardFooter className="flex justify-between">
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => router.push('/dataflows')}
                                >
                                    Cancel
                                </Button>
                                <Button type="submit" disabled={isSaving} variant={"secondary"}>
                                    {isSaving ? 'Creating...' : 'Create Dataflow'}
                                </Button>
                            </CardFooter>
                        </form>
                    </Card>
                )}
            </DashboardLayout>
        </AuthGuard>
    );
}