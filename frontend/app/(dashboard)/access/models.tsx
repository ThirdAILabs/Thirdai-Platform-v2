'use client';
import React, { useState, useEffect } from 'react';
import { Button } from '@mui/material';
import {
  fetchAllModels,
  updateModelAccessLevel,
  deleteModel,
  fetchAllTeams,
  fetchAllUsers,
} from '@/lib/backend';
import { UserContext } from '../../user_wrapper';
import ConditionalButton from '@/components/ui/ConditionalButton';

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
  // State variables
  const [models, setModels] = useState<Model[]>([]);
  const [teams, setTeams] = useState<Team[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [selectedType, setSelectedType] = useState<
    'Private Model' | 'Protected Model' | 'Public Model' | null
  >(null);
  const [selectedTeam, setSelectedTeam] = useState<string | null>(null);
  const { user } = React.useContext(UserContext);
  const modelPermissions: boolean[] = [];

  // Fetch models on component mount
  useEffect(() => {
    getModels();
  }, []);
  // Function to fetch models
  const getModels = async () => {
    try {
      const response = await fetchAllModels();
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
        })
      );
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
      console.log('teams data in user-> ', userData);
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
      const isConfirmed = window.confirm(
        `Are you sure you want to delete the model "${model.name}"?`
      );
      if (!isConfirmed) return;

      await deleteModel(model_identifier);
      await getModels();
    } catch (error) {
      console.error('Failed to delete model', error);
      alert('Failed to delete model: ' + error);
    }
  };

  //The function decides if the user is allowed to make changes in the access level or delete the model out of n-models.
  const getModelPermissions = () => {
    for (let index = 0; index < models.length; index++) {
      const model = models[index];
      if (user?.global_admin) {
        modelPermissions.push(true);
        continue;
      }
      if (model.owner === user?.username) modelPermissions.push(true);
      else {
        let value = false;
        if (user?.teams && model.type === 'Protected Model') {
          for (let itr = 0; itr < user?.teams.length; itr++) {
            if (user?.teams[itr].team_id === model.team && user.teams[itr].role === 'team_admin') {
              value = true;
              break;
            }
          }
          modelPermissions.push(value);
        } else modelPermissions.push(false);
      }
    }
  };
  getModelPermissions();

  return (
    <div className="mb-12">
      <h3 className="text-xl font-semibold text-gray-800 mb-4">Models</h3>
      <table className="w-full bg-white rounded-lg shadow-md overflow-hidden">
        <thead className="bg-gray-100">
          <tr>
            <th className="py-3 px-4 text-left text-gray-700">Model Name</th>
            <th className="py-3 px-4 text-left text-gray-700">Model Type</th>
            <th className="py-3 px-4 text-left text-gray-700">Access Details</th>
            <th className="py-3 px-4 text-left text-gray-700">Edit Model Access</th>
            <th className="py-3 px-4 text-left text-gray-700">Delete Model</th>
          </tr>
        </thead>
        <tbody>
          {models.map((model, index) => (
            <tr key={index} className="border-t">
              <td className="py-3 px-4 text-gray-800">{model.name}</td>
              <td className="py-3 px-4 text-gray-800">{model.type}</td>
              <td className="py-3 px-4 text-gray-800">
                {model.type === 'Private Model' && (
                  <div>
                    <div>Owner: {model.owner}</div>
                    <div>Users: {model.users?.join(', ') || 'None'}</div>
                  </div>
                )}
                {model.type === 'Protected Model' && (
                  <div>
                    <div>Owner: {model.owner}</div>
                    <div>Team: {teams.find((team) => team.id === model.team)?.name || 'None'}</div>
                  </div>
                )}
                {model.type === 'Public Model' && <div>Owner: {model.owner}</div>}
              </td>
              <td className="py-3 px-4">
                {editingIndex === index ? (
                  <div className="flex flex-col space-y-2">
                    <select
                      value={selectedType || model.type}
                      onChange={(e) =>
                        setSelectedType(
                          e.target.value as 'Private Model' | 'Protected Model' | 'Public Model'
                        )
                      }
                      className="border border-gray-300 rounded px-4 py-2"
                    >
                      <option value="Private Model">Private Model</option>
                      <option value="Protected Model">Protected Model</option>
                      <option value="Public Model">Public Model</option>
                    </select>

                    {selectedType === 'Protected Model' && (
                      <select
                        value={selectedTeam || ''}
                        onChange={(e) => setSelectedTeam(e.target.value)}
                        className="border border-gray-300 rounded px-4 py-2"
                      >
                        <option value="" disabled>
                          Select Team
                        </option>
                        {teams.map((team) => (
                          <option key={team.id} value={team.id}>
                            {team.name}
                          </option>
                        ))}
                      </select>
                    )}
                    <div className="flex space-x-2 mt-2">
                      <Button
                        onClick={() => handleModelTypeChange(index)}
                        className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
                      >
                        Confirm
                      </Button>
                      <Button
                        onClick={() => setEditingIndex(null)}
                        className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
                      >
                        Cancel
                      </Button>
                    </div>
                  </div>
                ) : (
                  <ConditionalButton
                    onClick={() => setEditingIndex(index)}
                    isDisabled={!modelPermissions[index]}
                    tooltipMessage="Global Admin, Model owner and Team Admin can change access"
                    variant="contained"
                  >
                    Change Access
                  </ConditionalButton>
                )}
              </td>

              <td className="py-3 px-4">
                <ConditionalButton
                  onClick={() => handleDeleteModel(index)}
                  isDisabled={!modelPermissions[index]}
                  tooltipMessage="Global Admin, Model owner and Team Admin can delete"
                  variant="contained"
                  color="error"
                >
                  Delete
                </ConditionalButton>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
