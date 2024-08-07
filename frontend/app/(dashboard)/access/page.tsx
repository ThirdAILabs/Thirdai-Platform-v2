import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from '@/components/ui/card';

export default function AccessPage() {
  const userRole = "Global Admin";
  const roleDescription = "This role has read and write access to all team members and models.";

  // Sample data for the models
  const models = [
    { name: 'Model A', type: 'Private Model', owner: 'Alice', users: ['Bob', 'Charlie'] },
    { name: 'Model B', type: 'Protected Model', owner: 'Alice', team: 'Team A', teamAdmin: 'Charlie' },
    { name: 'Model C', type: 'Public Model', owner: 'Bob' },
  ];

  // Sample data for the teams
  const teams = [
    { name: 'Team A', admin: 'Charlie', members: ['Alice', 'Bob', 'Charlie'] },
    { name: 'Team B', admin: 'Dave', members: ['Eve', 'Frank', 'Grace'] },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Manage Access</CardTitle>
        <CardDescription>View all personnel and their access.</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="mb-4">
          <h2 className="text-xl font-semibold">{userRole}</h2>
          <p>{roleDescription}</p>
        </div>

        {/* Models Section */}
        <div className="mb-8">
          <h3 className="text-lg font-semibold">Models</h3>
          <table className="min-w-full bg-white mb-8">
            <thead>
              <tr>
                <th className="py-2 px-4 text-left">Model Name</th>
                <th className="py-2 px-4 text-left">Model Type</th>
                <th className="py-2 px-4 text-left">Access Details</th>
              </tr>
            </thead>
            <tbody>
              {models.map((model, index) => (
                <tr key={index} className="border-t">
                  <td className="py-2 px-4">{model.name}</td>
                  <td className="py-2 px-4">{model.type}</td>
                  <td className="py-2 px-4">
                    {model.type === 'Private Model' && (
                      <div>
                        <div>Owner: {model.owner}</div>
                        <div>Users: {model.users?.join(', ') || 'None'}</div>
                      </div>
                    )}
                    {model.type === 'Protected Model' && (
                      <div>
                        <div>Owner: {model.owner}</div>
                        <div>Team: {model.team || 'None'}</div>
                        <div>Team Admin: {model.teamAdmin || 'None'}</div>
                      </div>
                    )}
                    {model.type === 'Public Model' && (
                      <div>
                        <div>Owner: {model.owner}</div>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Teams Section */}
        <div>
          <h3 className="text-lg font-semibold">Teams</h3>
          {teams.map((team, index) => (
            <div key={index} className="mb-8">
              <h4 className="text-md font-semibold">{team.name}</h4>
              <div className="mb-2">Admin: {team.admin}</div>
              <div className="mb-2">Members: {team.members.join(', ')}</div>
              <div>
                <h5 className="text-sm font-semibold">Protected Models</h5>
                <ul className="list-disc pl-5">
                  {models
                    .filter(model => model.type === 'Protected Model' && model.team === team.name)
                    .map((model, modelIndex) => (
                      <li key={modelIndex}>{model.name}</li>
                    ))}
                </ul>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
