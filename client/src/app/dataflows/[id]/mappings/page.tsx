'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/Tabs';
import DashboardLayout from '@/components/layout/DashboardLayout';
import PageHeader from '@/components/layout/PageHeader';
import { AuthGuard } from '@/contexts/AuthContext';
import api, { Dataflow, FieldMapping } from '@/lib/api';
import { formatDate, getStatusVariant, titleCase } from '@/lib/utils';
import toast from 'react-hot-toast';

// Field mapping form component
interface FieldMappingFormProps {
    mapping?: FieldMapping;
    dataflowId: number;
    onSave: () => void;
    onCancel: () => void;
}

function FieldMappingForm({ mapping, dataflowId, onSave, onCancel }: FieldMappingFormProps) {
    const [isLoading, setIsLoading] = useState(false);
    const [formData, setFormData] = useState<Partial<FieldMapping>>({
        dataflow_id: dataflowId,
        source_field: mapping?.source_field || '',
        dest_field: mapping?.dest_field || '',
        is_required: mapping?.is_required || false,
        default_value: mapping?.default_value || '',
        transform_type: mapping?.transform_type || 'none',
        transform_config: mapping?.transform_config || '{}'
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        const { name, value, type } = e.target;
        if (type === 'checkbox') {
            setFormData({ ...formData, [name]: (e.target as HTMLInputElement).checked });
        } else {
            setFormData({ ...formData, [name]: value });
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);

        try {
            if (mapping?.id) {
                // Update existing mapping
                await api.updateFieldMapping(dataflowId, mapping.id, formData);
                toast.success('Field mapping updated successfully');
            } else {
                // Create new mapping
                await api.createFieldMapping(dataflowId, formData);
                toast.success('Field mapping created successfully');
            }
            onSave();
        } catch (error) {
            console.error('Error saving field mapping:', error);
            toast.error('Failed to save field mapping');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle>{mapping ? 'Edit Field Mapping' : 'New Field Mapping'}</CardTitle>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <label htmlFor="source_field" className="text-sm font-medium">
                                Source Field
                            </label>
                            <input
                                id="source_field"
                                name="source_field"
                                value={formData.source_field}
                                onChange={handleChange}
                                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                placeholder="product.name"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <label htmlFor="dest_field" className="text-sm font-medium">
                                Destination Field
                            </label>
                            <input
                                id="dest_field"
                                name="dest_field"
                                value={formData.dest_field}
                                onChange={handleChange}
                                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                placeholder="product.title"
                                required
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <label htmlFor="transform_type" className="text-sm font-medium">
                                Transform Type
                            </label>
                            <select
                                id="transform_type"
                                name="transform_type"
                                value={formData.transform_type}
                                onChange={handleChange}
                                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            >
                                <option value="none">None (Direct Mapping)</option>
                                <option value="format">Format (e.g., Date)</option>
                                <option value="convert">Convert Type</option>
                                <option value="map">Map Values</option>
                                <option value="template">Template</option>
                            </select>
                        </div>
                        <div className="space-y-2">
                            <label htmlFor="default_value" className="text-sm font-medium">
                                Default Value
                            </label>
                            <input
                                id="default_value"
                                name="default_value"
                                value={formData.default_value}
                                onChange={handleChange}
                                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                placeholder="Default value if source field is empty"
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label htmlFor="transform_config" className="text-sm font-medium">
                            Transform Configuration (JSON)
                        </label>
                        <textarea
                            id="transform_config"
                            name="transform_config"
                            value={formData.transform_config}
                            onChange={handleChange}
                            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono"
                            rows={5}
                            placeholder='{"type": "string"}'
                        />
                    </div>

                    <div className="flex items-center">
                        <input
                            id="is_required"
                            name="is_required"
                            type="checkbox"
                            checked={formData.is_required}
                            onChange={handleChange}
                            className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                        />
                        <label htmlFor="is_required" className="ml-2 block text-sm text-gray-900">
                            Required Field
                        </label>
                    </div>

                    <div className="flex justify-end space-x-2">
                        <Button type="button" variant="outline" onClick={onCancel}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={isLoading}>
                            {isLoading ? 'Saving...' : 'Save Mapping'}
                        </Button>
                    </div>
                </form>
            </CardContent>
        </Card>
    );
}

export default function FieldMappingsPage() {
    const params = useParams();
    const dataflowId = parseInt(params.id as string);
    const router = useRouter();

    const [isLoading, setIsLoading] = useState(true);
    const [dataflow, setDataflow] = useState<Dataflow | null>(null);
    const [mappings, setMappings] = useState<FieldMapping[]>([]);
    const [showForm, setShowForm] = useState(false);
    const [editMapping, setEditMapping] = useState<FieldMapping | undefined>(undefined);
    const [activeTab, setActiveTab] = useState('mappings');

    // Example field recommendations based on data type
    const productFieldRecommendations = [
        { source: 'product.id', dest: 'product.id', description: 'Product ID' },
        { source: 'product.name', dest: 'product.title', description: 'Product name/title' },
        { source: 'product.description', dest: 'product.body_html', description: 'Product description' },
        { source: 'product.price[0].gross', dest: 'product.variants[0].price', description: 'Product price' },
        { source: 'product.stock', dest: 'product.variants[0].inventory_quantity', description: 'Stock quantity' }
    ];

    const orderFieldRecommendations = [
        { source: 'order.id', dest: 'order.id', description: 'Order ID' },
        { source: 'order.orderNumber', dest: 'order.name', description: 'Order number/name' },
        { source: 'order.customer.email', dest: 'order.email', description: 'Customer email' },
        { source: 'order.totalPrice', dest: 'order.total_price', description: 'Total order price' },
        { source: 'order.lineItems', dest: 'order.line_items', description: 'Order line items' }
    ];

    useEffect(() => {
        async function fetchData() {
            setIsLoading(true);
            try {
                // Fetch dataflow details
                const dataflowRes = await api.getDataflow(dataflowId);
                setDataflow(dataflowRes.data.data);

                // Fetch field mappings
                const mappingsRes = await api.getFieldMappings(dataflowId);
                setMappings(mappingsRes.data.data || []);
            } catch (error) {
                console.error('Error fetching data:', error);
                toast.error('Failed to load dataflow details');
                router.push('/dataflows');
            } finally {
                setIsLoading(false);
            }
        }

        fetchData();
    }, [dataflowId, router]);

    const handleAddMapping = () => {
        setEditMapping(undefined);
        setShowForm(true);
        setActiveTab('form');
    };

    const handleEditMapping = (mapping: FieldMapping) => {
        setEditMapping(mapping);
        setShowForm(true);
        setActiveTab('form');
    };

    const handleCloseForm = () => {
        setShowForm(false);
        setEditMapping(undefined);
        setActiveTab('mappings');
    };

    const handleSaveForm = async () => {
        setShowForm(false);
        setEditMapping(undefined);
        setActiveTab('mappings');

        // Refresh mappings
        try {
            const mappingsRes = await api.getFieldMappings(dataflowId);
            setMappings(mappingsRes.data.data || []);
        } catch (error) {
            console.error('Error refreshing mappings:', error);
        }
    };

    const handleDeleteMapping = async (mappingId: number) => {
        if (!window.confirm('Are you sure you want to delete this field mapping?')) {
            return;
        }

        try {
            await api.deleteFieldMapping(dataflowId, mappingId);
            toast.success('Field mapping deleted successfully');

            // Refresh mappings
            const mappingsRes = await api.getFieldMappings(dataflowId);
            setMappings(mappingsRes.data.data || []);
        } catch (error) {
            console.error('Error deleting mapping:', error);
            toast.error('Failed to delete field mapping');
        }
    };

    const handleAddRecommendedField = (source: string, dest: string) => {
        setEditMapping(undefined);
        setShowForm(true);
        setActiveTab('form');
    };

    if (isLoading) {
        return (
            <AuthGuard>
                <DashboardLayout>
                    <div className="flex justify-center items-center h-64">
                        <p>Loading field mappings...</p>
                    </div>
                </DashboardLayout>
            </AuthGuard>
        );
    }

    if (!dataflow) {
        return (
            <AuthGuard>
                <DashboardLayout>
                    <div className="flex justify-center items-center h-64">
                        <p>Dataflow not found</p>
                    </div>
                </DashboardLayout>
            </AuthGuard>
        );
    }

    const recommendations = dataflow.type === 'product'
        ? productFieldRecommendations
        : orderFieldRecommendations;

    return (
        <AuthGuard>
            <DashboardLayout>
                <PageHeader
                    title={`Field Mappings: ${dataflow.name}`}
                    description={`Configure field mappings for ${dataflow.type} data`}
                    actions={
                        <Button onClick={handleAddMapping}>
                            Add Field Mapping
                        </Button>
                    }
                />

                <div className="mb-6">
                    <Card>
                        <CardContent className="p-4">
                            <div className="flex flex-wrap gap-4 items-center">
                                <div>
                                    <span className="text-sm font-medium">Type:</span>{" "}
                                    <Badge variant="secondary">{dataflow.type}</Badge>
                                </div>
                                <div>
                                    <span className="text-sm font-medium">Status:</span>{" "}
                                    <Badge variant={getStatusVariant(dataflow.status)}>{dataflow.status}</Badge>
                                </div>
                                <div>
                                    <span className="text-sm font-medium">Source:</span>{" "}
                                    <span className="text-sm">{dataflow.source_connector.name} ({dataflow.source_connector.type})</span>
                                </div>
                                <div>
                                    <span className="text-sm font-medium">Destination:</span>{" "}
                                    <span className="text-sm">{dataflow.dest_connector.name} ({dataflow.dest_connector.type})</span>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                <Tabs value={activeTab} onValueChange={setActiveTab}>
                    <TabsList className="mb-4">
                        <TabsTrigger value="mappings">Field Mappings</TabsTrigger>
                        <TabsTrigger value="recommendations">Recommendations</TabsTrigger>
                        {showForm && <TabsTrigger value="form">Mapping Form</TabsTrigger>}
                    </TabsList>

                    <TabsContent value="mappings">
                        {mappings.length === 0 ? (
                            <Card>
                                <CardContent className="flex flex-col items-center justify-center py-12">
                                    <p className="text-muted-foreground mb-4">No field mappings defined yet</p>
                                    <Button onClick={handleAddMapping}>
                                        Add Your First Field Mapping
                                    </Button>
                                </CardContent>
                            </Card>
                        ) : (
                            <div className="grid gap-4">
                                {mappings.map((mapping) => (
                                    <Card key={mapping.id}>
                                        <CardContent className="p-4">
                                            <div className="flex flex-col md:flex-row md:items-center md:justify-between">
                                                <div className="mb-4 md:mb-0">
                                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-2">
                                                        <div>
                                                            <div className="flex items-center gap-2">
                                                                <span className="text-sm font-medium">Source Field:</span>
                                                                <code className="bg-gray-100 px-1 rounded text-sm">{mapping.source_field}</code>
                                                                {mapping.is_required && (
                                                                    <Badge variant="error">Required</Badge>
                                                                )}
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <div className="flex items-center gap-2">
                                                                <span className="text-sm font-medium">Destination Field:</span>
                                                                <code className="bg-gray-100 px-1 rounded text-sm">{mapping.dest_field}</code>
                                                            </div>
                                                        </div>
                                                        <div>
                                                            <span className="text-sm font-medium">Transform Type:</span>{" "}
                                                            <span className="text-sm">{titleCase(mapping.transform_type)}</span>
                                                        </div>
                                                        <div>
                                                            <span className="text-sm font-medium">Default Value:</span>{" "}
                                                            <span className="text-sm">{mapping.default_value || "—"}</span>
                                                        </div>
                                                    </div>
                                                </div>
                                                <div className="flex space-x-2">
                                                    <Button
                                                        variant="outline"
                                                        size="sm"
                                                        onClick={() => handleEditMapping(mapping)}
                                                    >
                                                        Edit
                                                    </Button>
                                                    <Button
                                                        variant="destructive"
                                                        size="sm"
                                                        onClick={() => handleDeleteMapping(mapping.id)}
                                                    >
                                                        Delete
                                                    </Button>
                                                </div>
                                            </div>
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>
                        )}
                    </TabsContent>

                    <TabsContent value="recommendations">
                        <Card>
                            <CardHeader>
                                <CardTitle>Recommended Field Mappings for {titleCase(dataflow.type)} Data</CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="grid gap-4">
                                    {recommendations.map((rec, index) => (
                                        <div key={index} className="flex items-center justify-between border-b pb-2">
                                            <div>
                                                <div className="flex items-center gap-2">
                                                    <code className="bg-gray-100 px-1 rounded text-sm">{rec.source}</code>
                                                    <span className="text-gray-500">→</span>
                                                    <code className="bg-gray-100 px-1 rounded text-sm">{rec.dest}</code>
                                                </div>
                                                <p className="text-sm text-muted-foreground mt-1">{rec.description}</p>
                                            </div>
                                            <Button
                                                size="sm"
                                                onClick={() => handleAddRecommendedField(rec.source, rec.dest)}
                                            >
                                                Add Mapping
                                            </Button>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    </TabsContent>

                    {showForm && (
                        <TabsContent value="form">
                            <FieldMappingForm
                                mapping={editMapping}
                                dataflowId={dataflowId}
                                onSave={handleSaveForm}
                                onCancel={handleCloseForm}
                            />
                        </TabsContent>
                    )}
                </Tabs>
            </DashboardLayout>
        </AuthGuard>
    );
}