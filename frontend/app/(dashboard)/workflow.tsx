import Link from 'next/link';
import { AlertCircle } from 'lucide-react';
import { useContext, useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@mui/material';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { MoreHorizontal } from 'lucide-react';
import { TableCell, TableRow } from '@/components/ui/table';
import {
  Workflow,
  start_workflow,
  stop_workflow,
  delete_workflow,
  getTrainingStatus,
  getDeployStatus,
  getTrainingLogs,
  getDeploymentLogs,
} from '@/lib/backend';
import { Modal } from '@/components/ui/Modal';
import { InformationCircleIcon } from '@heroicons/react/solid';
import { Model, getModels } from '@/utils/apiRequests';
import { UserContext } from '../user_wrapper';

enum DeployStatus {
  None = '',
  TrainingFailed = 'Training failed',
  Training = 'Training',
  Inactive = 'Inactive',
  Starting = 'Starting',
  Active = 'Active',
  Failed = 'Failed',
}

interface ErrorState {
  type: 'training' | 'deployment';
  messages: string[];
}

export function WorkFlow({ workflow }: { workflow: Workflow }) {
  const { user } = useContext(UserContext);
  const [deployStatus, setDeployStatus] = useState<DeployStatus>(DeployStatus.None);
  const [deployType, setDeployType] = useState<string>('');
  const [modelOwner, setModelOwner] = useState<{ [key: string]: string }>({});

  useEffect(() => {
    getModelsData();
  }, []);

  async function getModelsData() {
    const modelData = await getModels();
    const tempModelOwner: { [key: string]: string } = {}; // TypeScript object to store name as key and owner as value
    if (modelData) {
      for (let index = 0; index < modelData.length; index++) {
        const name = modelData[index].name;
        const owner = modelData[index].owner;
        tempModelOwner[name] = owner;
      }
    }
    setModelOwner(tempModelOwner);
  }
  function goToEndpoint() {
    switch (workflow.type) {
      case 'enterprise-search': {
        // enterprise-search is rag with generation

        // TODO don't use url params
        const llmProvider = `${workflow.attributes.llm_provider}`;
        const ifGenerationOn = true;
        const chatMode = workflow.attributes.default_mode == 'chat';
        const newUrl = `/semantic-search/${workflow.model_id}?workflowId=${workflow.model_id}&ifGenerationOn=${ifGenerationOn}&genAiProvider=${llmProvider}&chatMode=${chatMode}`;
        window.open(newUrl, '_blank');
        break;
      }
      case 'ndb': {
        // ndb is is rag without generation
        const llmProvider = workflow.attributes.llm_provider
          ? workflow.attributes.llm_provider
          : null;
        // TODO don't use url params
        const ifGenerationOn = llmProvider != null;
        const newUrl = `/semantic-search/${workflow.model_id}?workflowId=${workflow.model_id}&ifGenerationOn=${ifGenerationOn}&genAiProvider=${llmProvider}`;
        window.open(newUrl, '_blank');
        break;
      }
      case 'udt': {
        const prefix =
          workflow.sub_type === 'token' ? '/token-classification' : '/text-classification';
        window.open(`${prefix}/${workflow.model_id}`, '_blank');
        break;
      }
      default:
        throw new Error(`Invalid workflow type ${workflow.type}`);
        break;
    }
  }

  function getButtonValue(status: DeployStatus): string {
    switch (status) {
      case DeployStatus.TrainingFailed:
        return 'Training failed';
      case DeployStatus.Training:
        return 'Training...';
      case DeployStatus.Inactive:
        return 'Start';
      case DeployStatus.Starting:
        return 'Starting...';
      case DeployStatus.Active:
        return 'Endpoint';
      case DeployStatus.Failed:
        return 'Failed';
      default:
        return '-';
    }
  }

  const handleDeploy = async () => {
    if (deployStatus == DeployStatus.Inactive) {
      setDeployStatus(DeployStatus.Starting);
      try {
        await start_workflow(workflow.username, workflow.model_name);
      } catch (e) {
        console.error('Failed to start workflow.', e);
      }
    }
  };

  useEffect(() => {
    if (workflow.train_status === 'failed') {
      setDeployStatus(DeployStatus.TrainingFailed);
    } else if (workflow.train_status !== 'complete') {
      setDeployStatus(DeployStatus.Training);
    } else if (workflow.deploy_status === 'failed') {
      setDeployStatus(DeployStatus.Failed);
    } else if (workflow.deploy_status === 'starting') {
      setDeployStatus(DeployStatus.Starting);
    } else if (workflow.deploy_status === 'not_started' || workflow.deploy_status === 'stopped') {
      setDeployStatus(DeployStatus.Inactive);
    } else if (workflow.deploy_status === 'complete') {
      setDeployStatus(DeployStatus.Active);
    }
  }, [workflow.train_status, workflow.deploy_status, deployStatus]);

  useEffect(() => {
    if (workflow.type === 'ndb') {
      if (workflow.attributes.default_mode && workflow.attributes.default_mode == 'chat') {
        setDeployType('Chatbot');
      } else {
        setDeployType('Enterprise Search');
      }
    } else if (workflow.type === 'udt') {
      setDeployType('Natural Language Processing');
    } else if (workflow.type === 'enterprise-search') {
      setDeployType('Enterprise Search & Summarizer');
    }
  }, [workflow.type]);

  const getBadgeColor = (status: DeployStatus) => {
    switch (status) {
      case DeployStatus.Active:
        return 'bg-green-500 text-white'; // Green for good status
      case DeployStatus.Starting:
        return 'bg-yellow-500 text-white'; // Yellow for in-progress status
      case DeployStatus.Inactive:
        return 'bg-gray-500 text-white'; // Gray for inactive status
      case DeployStatus.Training:
        return 'bg-blue-500 text-white';
      case DeployStatus.TrainingFailed: // New case for training failed
        return 'bg-red-500 text-white';
      case DeployStatus.Failed:
        return 'bg-red-500 text-white'; // Red for error statuses
      default:
        return 'bg-gray-500 text-white'; // Default to gray if status is unknown
    }
  };

  const [showModal, setShowModal] = useState(false);

  const toggleModal = () => {
    setShowModal(!showModal);
  };

  const formatBytesToMB = (bytes: string) => {
    return (parseInt(bytes) / (1024 * 1024)).toFixed(2) + ' MB';
  };

  // Add new state for error handling
  const [error, setError] = useState<ErrorState | null>(null);
  const [showErrorModal, setShowErrorModal] = useState(false);

  useEffect(() => {
    const fetchStatuses = async () => {
      try {
        if (workflow.username && workflow.model_name) {
          const modelIdentifier = `${workflow.username}/${workflow.model_name}`;
          const [trainStatus, deployStatus] = await Promise.all([
            getTrainingStatus(modelIdentifier),
            getDeployStatus(modelIdentifier),
          ]);

          // Check training status first
          if (trainStatus.data.train_status === 'failed' && trainStatus.data.messages?.length > 0) {
            setError({
              type: 'training',
              messages: trainStatus.data.messages,
            });
            return; // Exit early if training failed
          }

          // Only check deployment if training was successful
          if (
            deployStatus.data.deploy_status === 'failed' &&
            deployStatus.data.messages?.length > 0
          ) {
            setError({
              type: 'deployment',
              messages: deployStatus.data.messages,
            });
          } else {
            setError(null); // Clear error if everything is successful
          }
        }
      } catch (error) {
        console.error('Error fetching statuses:', error);
      }
    };

    // Initial fetch
    fetchStatuses();

    // Set up polling interval
    const intervalId = setInterval(fetchStatuses, 2000);

    // Cleanup function to clear interval when component unmounts
    return () => clearInterval(intervalId);
  }, [workflow.username, workflow.model_name]); // Dependencies stay the same

  return (
    <TableRow>
      <TableCell className="font-bold text-center">{workflow.model_name}</TableCell>
      <TableCell className="text-center font-medium">
        <Badge variant="outline" className={`capitalize ${getBadgeColor(deployStatus)}`}>
          {deployStatus}
        </Badge>
      </TableCell>
      <TableCell className="hidden md:table-cell text-center font-medium">{deployType}</TableCell>
      <TableCell className="hidden md:table-cell text-center font-medium">
        {new Date(workflow.publish_date).toLocaleDateString('en-US', {
          year: 'numeric',
          month: 'long',
          day: 'numeric',
        })}
      </TableCell>
      <TableCell className="hidden md:table-cell text-center font-medium">
        <Button
          onClick={deployStatus === 'Active' ? goToEndpoint : handleDeploy}
          variant="contained"
          style={{ width: '100px' }}
          disabled={deployStatus != DeployStatus.Active && deployStatus != DeployStatus.Inactive}
        >
          {getButtonValue(deployStatus)}
        </Button>
      </TableCell>
      <TableCell className="text-center font-medium">
        <button onClick={toggleModal} className="text-gray-400 hover:text-gray-600 text-sm">
          <InformationCircleIcon className="h-6 w-6" />
        </button>
      </TableCell>
      <TableCell className="text-center font-medium">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              aria-haspopup="true"
              size="small"
              variant="text" // Using "text" as base variant
              sx={{
                color: 'inherit', // Default text color
                '&:hover': {
                  backgroundColor: 'var(--accent)', // Replace with your accent color
                  color: 'var(--accent-foreground)', // Replace with your foreground color for hover
                },
              }}
            >
              <MoreHorizontal className="h-4 w-4" />
              <span className="sr-only">Toggle menu</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            {deployStatus === DeployStatus.Active && (
              <DropdownMenuItem>
                <form>
                  <button
                    type="button"
                    onClick={async () => {
                      try {
                        const response = await stop_workflow(
                          workflow.username,
                          workflow.model_name
                        );
                        console.log('Workflow undeployed successfully:', response);
                        // Optionally, update the UI state to reflect the undeployment
                        setDeployStatus(DeployStatus.Inactive);
                      } catch (error) {
                        console.error('Error undeploying workflow:', error);
                        alert('Error undeploying workflow:' + error);
                      }
                    }}
                  >
                    Stop App
                  </button>
                </form>
              </DropdownMenuItem>
            )}

            {workflow.type === 'ndb' &&
              (modelOwner[workflow.model_name] === user?.username || user?.global_admin) && (
                <DropdownMenuItem>
                  <form>
                    <button
                      type="button"
                      onClick={async () => {
                        if (window.confirm('Are you sure you want to delete this workflow?')) {
                          try {
                            const response = await delete_workflow(
                              workflow.username,
                              workflow.model_name
                            );
                            console.log('Workflow deleted successfully:', response);
                          } catch (error) {
                            console.error('Error deleting workflow:', error);
                            alert('Error deleting workflow:' + error);
                          }
                        }
                      }}
                    >
                      Delete App
                    </button>
                  </form>
                </DropdownMenuItem>
              )}

            {workflow.type === 'ndb' &&
              (modelOwner[workflow.model_name] === user?.username || user?.global_admin) && (
                <Link
                  href={`/analytics?id=${encodeURIComponent(workflow.model_id)}&username=${encodeURIComponent(workflow.username)}&model_name=${encodeURIComponent(workflow.model_name)}&old_model_id=${encodeURIComponent(workflow.model_id)}`}
                >
                  <DropdownMenuItem>
                    <button type="button">Search usage stats</button>
                  </DropdownMenuItem>
                </Link>
              )}

            {workflow.type === 'udt' && (
              <Link
                href={`/analytics?id=${encodeURIComponent(workflow.model_id)}&username=${encodeURIComponent(workflow.username)}&model_name=${encodeURIComponent(workflow.model_name)}&old_model_id=${encodeURIComponent(workflow.model_id)}`}
              >
                <DropdownMenuItem>
                  <button type="button">NLP usage stats</button>
                </DropdownMenuItem>
              </Link>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>

      {/* Add error notification icon in the last cell */}
      <TableCell className="text-right pr-4">
        {error && (
          <button
            onClick={() => setShowErrorModal(true)}
            className="inline-flex items-center justify-center rounded-full bg-red-100 p-2 hover:bg-red-200 transition-colors"
          >
            <AlertCircle className="h-5 w-5 text-red-600" />
          </button>
        )}
      </TableCell>

      {/* Error Modal */}
      {showErrorModal && error && (
        <Modal onClose={() => setShowErrorModal(false)}>
          <div className="p-6 max-h-[80vh] flex flex-col">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">
                {error.type === 'training' ? 'Training Failed' : 'Deployment Failed'}
              </h2>
              <div className="flex space-x-2">
                <button
                  onClick={() => {
                    const errorText = error.messages.join('\n');
                    navigator.clipboard.writeText(errorText).then(() => {
                      const notification = document.createElement('div');
                      notification.className =
                        'fixed bottom-4 right-4 bg-gray-800 text-white px-4 py-2 rounded-md shadow-lg';
                      notification.textContent = 'Error copied to clipboard';
                      document.body.appendChild(notification);
                      setTimeout(() => {
                        document.body.removeChild(notification);
                      }, 2000);
                    });
                  }}
                  className="inline-flex items-center px-3 py-1 space-x-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-md text-gray-700"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                    />
                  </svg>
                  <span>Copy Error</span>
                </button>

                <button
                  onClick={async () => {
                    try {
                      const modelIdentifier = `${workflow.username}/${workflow.model_name}`;
                      const logs = await (error.type === 'training'
                        ? getTrainingLogs(modelIdentifier)
                        : getDeploymentLogs(modelIdentifier));

                      console.log('logs', logs);

                      // Create base content with metadata
                      const contentParts = [
                        `Error Type: ${error.type}`,
                        `Time: ${new Date().toISOString()}`,
                        `Model: ${modelIdentifier}`,
                        '',
                        'Error Messages:',
                        error.messages.join('\n'),
                        '',
                      ];

                      // Add each log entry with index
                      logs.data.forEach((log, index) => {
                        contentParts.push(
                          `Log Entry ${index + 1}:`,
                          '----------------',
                          'Standard Output:',
                          log.stdout,
                          '',
                          'Standard Error:',
                          log.stderr,
                          '' // Add empty line between log entries
                        );
                      });

                      const content = contentParts.join('\n').trim();

                      const blob = new Blob([content], { type: 'text/plain' });
                      const url = window.URL.createObjectURL(blob);
                      const a = document.createElement('a');
                      a.href = url;
                      a.download = `${error.type}_logs_${workflow.model_name}_${new Date().toISOString()}.txt`;
                      document.body.appendChild(a);
                      a.click();
                      window.URL.revokeObjectURL(url);
                      document.body.removeChild(a);

                      const notification = document.createElement('div');
                      notification.className =
                        'fixed bottom-4 right-4 bg-gray-800 text-white px-4 py-2 rounded-md shadow-lg';
                      notification.textContent = 'Logs downloaded successfully';
                      document.body.appendChild(notification);
                      setTimeout(() => {
                        document.body.removeChild(notification);
                      }, 2000);
                    } catch (err) {
                      console.error('Failed to download logs:', err);
                      const notification = document.createElement('div');
                      notification.className =
                        'fixed bottom-4 right-4 bg-red-600 text-white px-4 py-2 rounded-md shadow-lg';
                      notification.textContent = 'Failed to download logs';
                      document.body.appendChild(notification);
                      setTimeout(() => {
                        document.body.removeChild(notification);
                      }, 2000);
                    }
                  }}
                  className="inline-flex items-center px-3 py-1 space-x-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-md text-gray-700"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                    />
                  </svg>
                  <span>Download Logs</span>
                </button>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto min-h-0">
              <div className="space-y-2">
                <h3 className="font-medium text-gray-700 sticky top-0 bg-white py-2">
                  Error Details:
                </h3>
                <ul className="list-disc pl-5 space-y-2">
                  {error.messages.map((message, index) => (
                    <li key={index} className="text-gray-600">
                      {message}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
            <div className="mt-6 flex justify-end pt-4 border-t sticky bottom-0 bg-white">
              <button
                onClick={() => setShowErrorModal(false)}
                className="px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-md text-gray-700 font-medium"
              >
                Close
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* Modal for displaying model details */}
      {showModal && (
        <Modal onClose={toggleModal}>
          <div className="p-4">
            <h2 className="text-lg font-bold mb-4">App Details</h2>
            <div className="overflow-x-auto">
              <table className="min-w-full table-auto border-collapse border border-gray-300">
                <thead>
                  <tr className="bg-gray-200">
                    <th className="border px-4 py-2 text-left">Model Name</th>
                    <th className="border px-4 py-2 text-left">Size on Disk (MB)</th>
                    <th className="border px-4 py-2 text-left">Size in Memory (MB)</th>
                  </tr>
                </thead>
                <tbody>
                  {/* {workflow.models.map((model, index) => (
                    <tr key={index} className="hover:bg-gray-100">
                      <td className="border px-4 py-2">{model.model_name}</td>
                      <td className="border px-4 py-2">{formatBytesToMB(model.size)}</td>
                      <td className="border px-4 py-2">{formatBytesToMB(model.size_in_memory)}</td>
                    </tr>
                  ))} */}
                  {/* {workflow.models.map((model, index) => ( */}
                  <tr className="hover:bg-gray-100">
                    <td className="border px-4 py-2">{workflow.model_name}</td>
                    <td className="border px-4 py-2">{formatBytesToMB(workflow.size)}</td>
                    <td className="border px-4 py-2">{formatBytesToMB(workflow.size_in_memory)}</td>
                  </tr>
                  {/* ))} */}
                </tbody>
              </table>
            </div>
          </div>
        </Modal>
      )}
    </TableRow>
  );
}
