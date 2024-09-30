import React, { useState, useEffect } from 'react';
import { Button, TextField } from '@mui/material';
import {
    fetchAllModels,
    updateModelAccessLevel,
    deleteModel,
    fetchAllTeams,
    fetchAllUsers
} from '@/lib/backend';
import { UserContext } from '../../user_wrapper';

// Define types
type Model = {
    name: string;
    type: 'Private Model' | 'Protected Model' | 'Public Model';
    owner: string;
    users?: string[];
    team?: string;
    teamAdmin?: string;
    domain: string;
    latency: string;
    modelId: string;
    numParams: string;
    publishDate: string;
    size: string;
    sizeInMemory: string;
    subType: string;
    thirdaiVersion: string;
    trainingTime: string;
};

type Team = {
    id: string;
    name: string;
    admins: string[];
    members: string[];
};

type User = {
    id: string;
    name: string;
    email: string;
    role: 'Member' | 'Team Admin' | 'Global Admin';
    teams: { id: string; name: string; role: 'Member' | 'team_admin' | 'Global Admin' }[];
};

export default function Models() {
    const { user } = React.useContext(UserContext);

    // State variables
    const [models, setModels] = useState<Model[]>([]);
    const [teams, setTeams] = useState<Team[]>([]);
    const [users, setUsers] = useState<User[]>([]);
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [selectedType, setSelectedType] = useState<'Private Model' | 'Protected Model' | 'Public Model' | null>(null);
    const [selectedTeam, setSelectedTeam] = useState<string | null>(null);

    // Fetch models on component mount
    useEffect(() => {
        getModels();
    }, []);

    // Function to fetch models
    const getModels = async () => {
        try {
            const response = await fetchAllModels();
            const modelData = response.data.map((model): Model => ({
                name: model.model_name,
                type: model.access_level === 'private' ? 'Private Model' :
                    model.access_level === 'protected' ? 'Protected Model' : 'Public Model',
                owner: model.username,
                users: [],
                team: model.team_id !== 'None' ? model.team_id : undefined,
                teamAdmin: undefined,
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
            }));
            setModels(modelData);
            console.log('Fetched Models:', modelData);
        } catch (error) {
            console.error('Failed to fetch models', error);
            alert('Failed to fetch models' + error);
        }
    };

    useEffect(() => {
        getUsers();
    }, []);

    const getUsers = async () => {
        try {
            const response = await fetchAllUsers();
            console.log('Fetched Users:', response.data);
            const userData = response.data.map(
                (user): User => ({
                    id: user.id,
                    name: user.username,
                    email: user.email,
                    role: user.global_admin ? 'Global Admin' : 'Member',
                    teams: user.teams.map((team) => ({
                        id: team.team_id,
                        name: team.team_name,
                        role: team.role,
                    })),

                })
            );
            setUsers(userData);
        } catch (error) {
            console.error('Failed to fetch users', error);
            alert('Failed to fetch users' + error);
        }
    };

    useEffect(() => {
        getTeams();
    }, [users]);
    const getTeams = async () => {
        try {
            const response = await fetchAllTeams();
            const teamData = response.data.map((team): Team => {
                const members: string[] = [];
                const admins: string[] = [];

                users.forEach((user) => {
                    const userTeam = user.teams.find((ut) => ut.id === team.id);
                    if (userTeam) {
                        members.push(user.name);
                        if (userTeam.role === 'team_admin') {
                            admins.push(user.name);
                        }
                    }
                });

                return {
                    id: team.id,
                    name: team.name,
                    admins: admins,
                    members: members,
                };
            });

            setTeams(teamData);
        } catch (error) {
            console.error('Failed to fetch teams', error);
            alert('Failed to fetch teams' + error);
        }
    };
    // Function to handle model type change
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
            }
            team_id = teams.find(team => team.name === selectedTeam)?.id;
            await updateModelAccessLevel(model_identifier, access_level, team_id);
            await getModels();

            setEditingIndex(null);
            setSelectedType(null);
            setSelectedTeam(null);
        } catch (error) {
            console.error('Failed to update model access level', error);
            alert('Failed to update model access level' + error);
        }
    };

    // Function to handle model deletion
    const handleDeleteModel = async (index: number) => {
        const model = models[index];
        const model_identifier = `${model.owner}/${model.name}`;

        try {
            const isConfirmed = window.confirm(`Are you sure you want to delete the model "${model.name}"?`);
            if (!isConfirmed) return;

            await deleteModel(model_identifier);
            await getModels();
        } catch (error) {
            console.error('Failed to delete model', error);
            alert('Failed to delete model: ' + error);
        }
    };
    return (
        <div className="mb-12">
            <h3 className="text-xl font-semibold text-gray-800 mb-4">Models</h3>
            <table className="min-w-full bg-white border border-gray-300">
                <thead>
                    <tr className="bg-gray-100">
                        <th className="py-2 px-4 border-b">Name</th>
                        <th className="py-2 px-4 border-b">Type</th>
                        <th className="py-2 px-4 border-b">Owner</th>
                        <th className="py-2 px-4 border-b">Team</th>
                        <th className="py-2 px-4 border-b">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {models.map((model, index) => (
                        <tr key={model.modelId} className="hover:bg-gray-50">
                            <td className="py-2 px-4 border-b">{model.name}</td>
                            <td className="py-2 px-4 border-b">
                                {editingIndex === index ? (
                                    <select
                                        value={selectedType || model.type}
                                        onChange={(e) => setSelectedType(e.target.value as any)}
                                        className="border rounded px-2 py-1"
                                    >
                                        <option value="Private Model">Private Model</option>
                                        <option value="Protected Model">Protected Model</option>
                                        <option value="Public Model">Public Model</option>
                                    </select>
                                ) : (
                                    model.type
                                )}
                            </td>
                            <td className="py-2 px-4 border-b">{model.owner}</td>
                            <td className="py-2 px-4 border-b">
                                {editingIndex === index && selectedType === 'Protected Model' ? (
                                    <TextField
                                        value={selectedTeam || ''}
                                        onChange={(e) => setSelectedTeam(e.target.value)}
                                        placeholder="Enter team ID"
                                        size="small"
                                    />
                                ) : (
                                    model.team || 'N/A'
                                )}
                            </td>
                            <td className="py-2 px-4 border-b">
                                {editingIndex === index ? (
                                    <>
                                        <Button
                                            onClick={() => handleModelTypeChange(index)}
                                            variant="contained"
                                            color="primary"
                                            size="small"
                                            className="mr-2"
                                        >
                                            Save
                                        </Button>
                                        <Button
                                            onClick={() => setEditingIndex(null)}
                                            variant="outlined"
                                            color="secondary"
                                            size="small"
                                        >
                                            Cancel
                                        </Button>
                                    </>
                                ) : (
                                    <>
                                        <Button
                                            onClick={() => setEditingIndex(index)}
                                            variant="outlined"
                                            color="primary"
                                            size="small"
                                            className="mr-2"
                                        >
                                            Edit
                                        </Button>
                                        <Button
                                            onClick={() => handleDeleteModel(index)}
                                            variant="outlined"
                                            color="secondary"
                                            size="small"
                                        >
                                            Delete
                                        </Button>
                                    </>
                                )}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}