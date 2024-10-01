'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import Models from './models';
import Workflows from './workflows';
import Teams from './teams';
import Users from './users';
import OpenAIKey from './apiKey';
import { UserContext } from '../../user_wrapper';
import { usePathname, useSearchParams } from 'next/navigation';
export default function AccessPage() {
  const { user } = React.useContext(UserContext);
  console.log("page.tsx user", user);
  const [activeTab, setActiveTab] = useState('models');
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    // Get the active tab from the URL search params on initial load
    const tab = searchParams.get('tab');
    if (tab && ['models', 'workflows', 'teams', 'users', 'openai'].includes(tab)) {
      setActiveTab(tab);
    }
  }, [searchParams]);

  const handleTabChange = (value: string) => {
    setActiveTab(value);
    // Update the URL search params when changing tabs
    const newSearchParams = new URLSearchParams(searchParams);
    newSearchParams.set('tab', value);
    window.history.pushState(null, '', `${pathname}?${newSearchParams.toString()}`);
  };

  //Determine the user role
  let userRole = '';
  let roleDescription = '';

  if (user?.global_admin) {
    userRole = 'Global Admin';
    roleDescription = 'This role has read and write access to all members, models, and applications.';
  } else if (user?.teams.some((team) => team.role === 'team_admin')) {
    userRole = 'Team Admin';
    roleDescription = 'This role has read and write access to all team members, models, and applications in the team.';
  } else {
    userRole = 'User';
    roleDescription = 'This role has limited access based on specific team permissions.';
  }

  return (
    <div className="max-w-7xl mx-auto p-6 bg-gray-50 min-h-screen">
      <Card className="shadow-lg">
        <CardHeader className="bg-blue-500 text-white p-6 rounded-t-lg">
          <CardTitle className="text-2xl font-bold">Manage Access</CardTitle>
          <CardDescription className="text-white">
            View all personnel and their access.
          </CardDescription>
        </CardHeader>
        <CardContent className="p-6 bg-white rounded-b-lg">
          <div className="mb-8">
            <h2 className="text-2xl font-semibold text-gray-800">{userRole}</h2>
            <p className="text-gray-600">{roleDescription}</p>
          </div>

          <Tabs value={activeTab} onValueChange={handleTabChange}>
            <TabsList>
              <TabsTrigger value="models">Models</TabsTrigger>
              <TabsTrigger value="workflows">Workflows</TabsTrigger>
              <TabsTrigger value="teams">Teams</TabsTrigger>
              <TabsTrigger value="users">Users</TabsTrigger>
              <TabsTrigger value="openai">API Key</TabsTrigger>
            </TabsList>
            <TabsContent value="models">
              <Models />
            </TabsContent>
            <TabsContent value="workflows">
              <Workflows />
            </TabsContent>
            <TabsContent value="teams">
              <Teams />
            </TabsContent>
            <TabsContent value="users">
              <Users />
            </TabsContent>
            <TabsContent value="openai">
              <OpenAIKey />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
