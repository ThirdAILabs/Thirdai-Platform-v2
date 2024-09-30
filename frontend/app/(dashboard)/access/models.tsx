// Models.tsx
import React, { useState, useEffect } from 'react';
// Import necessary components and functions

export default function Models() {

    const [models, setModels] = useState<Model[]>([]);
    const [selectedType, setSelectedType] = useState<
        'Private Model' | 'Protected Model' | 'Public Model' | null
    >(null);
    const getModels = async () => {
        try {
            const response = await fetchAllModels();
            console.log('Fetched Models:', response.data); // Print out the results
            const modelData = response.data.map(
                (model): Model => ({
                    name: model.model_name,
                    type:
                        model.access_level === 'private'
                            ? 'Private Model'
                            : model.access_level === 'protected'
                                ? 'Protected Model'
                                : 'Public Model',
                    owner: model.username,
                    users: [], // To be populated later
                    team: model.team_id !== 'None' ? model.team_id : undefined,
                    teamAdmin: undefined, // To be populated later
                    domain: model.domain,
                    latency: model.latency,
                    modelId: model.model_id,
                    numParams: model.num_params,
                    publishDate: model.publish_date,
                    size: model.size,
                    sizeInMemory: model.size_in_memory,
                    subType: model.sub_type,
                    thirdaiVersion: model.thirdai_version,
                    trainingTime: model.training_time,
                })
            );
            setModels(modelData);
        } catch (error) {
            console.error('Failed to fetch models', error);
            alert('Failed to fetch models' + error);
        }
    };
    const handleModelTypeChange = async (index: number) => {
        if (!selectedType) return;

        try {
            const model = models[index];
            const model_identifier = `${model.owner}/${model.name}`;
            let access_level: 'private' | 'protected' | 'public' = 'private';
            let team_id: string | undefined;

            switch (selectedType) {
                case 'Private Model':
                    access_level = 'private';
                    break;
                case 'Protected Model':
                    access_level = 'protected';
                    team_id = selectedTeam || undefined;
                    break;
                case 'Public Model':
                    access_level = 'public';
                    break;
                default:
                    return;
            }

            // Call the API to update the model access level
            await updateModelAccessLevel(model_identifier, access_level, team_id);

            // Update the models state
            await getModels();
            await getUsers();
            await getTeams();

            // Reset editing state
            setEditingIndex(null);
            setSelectedType(null);
            setSelectedTeam(null);
        } catch (error) {
            console.error('Failed to update model access level', error);
            alert('Failed to update model access level' + error);
        }
    };
    const handleDeleteModel = async (index: number) => {
        const model = models[index];
        const model_identifier = `${model.owner}/${model.name}`;

        try {
            // Confirm deletion with the user
            const isConfirmed = window.confirm(
                `Are you sure you want to delete the model "${model.name}"?`
            );
            if (!isConfirmed) return;

            // Call the API to delete the model
            await deleteModel(model_identifier);

            // Optionally, refresh the models after deletion
            await getModels();
            await getUsers();
            await getTeams();
        } catch (error) {
            console.error('Failed to delete model', error);
            alert('Failed to delete model: ' + error);
        }
    };
    // Copy the models-related state and functions from the original file
    // This includes getModels, handleModelTypeChange, handleDeleteModel, etc.

    return (
        <div className="mb-12">
            <h3 className="text-xl font-semibold text-gray-800 mb-4">Models</h3>
            {/* Copy the models table and related JSX from the original file */}
        </div>
    );
}