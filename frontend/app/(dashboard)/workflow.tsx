import Link from 'next/link';
import { useEffect, useState } from 'react';

import Image from 'next/image';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { MoreHorizontal } from 'lucide-react';
import { TableCell, TableRow } from '@/components/ui/table';
import { deleteModel } from './actions';
import { deployModel, getDeployStatus, stopDeploy, getAccessToken, deploymentBaseUrl, listDeployments, 
          Workflow, validate_workflow, start_workflow } from '@/lib/backend';
import { useRouter } from 'next/navigation';

export function WorkFlow({ workflow, pending }: { workflow: Workflow, pending?: boolean }) {
  const router = useRouter();
  const [modelIdentifier, setModelIdentifier] = useState<string>('');
  const [deployStatus, setDeployStatus] = useState<string>(pending ? 'in queue' :'');
  const [deployType, setDeployType] = useState<string>('');
  const [deploymentIdentifier, setDeploymentIdentifier] = useState<string | null>(null);
  const [nerRAGEndpoint, setNerRAGEndpoint] = useState<string | null>(null);

  // useEffect(() => {
  //   const username = model.username;
  //   const modelIdentifier = `${username}/${model.model_name}`;
  //   setDeploymentIdentifier(`${modelIdentifier}:${username}/${model.model_name}`);
  //   setModelIdentifier(modelIdentifier);
  // }, [])

  // useEffect(() => {
  //   if (deploymentIdentifier) {
  //     const fetchDeployStatus = () => {

  //       if (model.type === 'rag') {
  //         let ndbModelId = model.ndb_model_id
  //         let tokenModelId = model.token_model_id

  //         if (ndbModelId) {
  //           getDeployStatus({ deployment_identifier: `${ndbModelId}:${ndbModelId}`, model_identifier: modelIdentifier })
  //           .then((response) => {
  //             // console.log('Deployment status response:', response);
  //             if (response.data.model_id && response.data.deploy_status === 'complete') {
                
  //               // console.log('The NDB model is already deployed, and deployment ID is: ', response.data.model_id)
  //               setDeploymentId(response.data.model_id)

  //               // check if NER model is deployed
  //               if (! tokenModelId) {
  //                 setDeployStatus('Deployed')
  //               } else {
  //                 getDeployStatus({ deployment_identifier: `${tokenModelId}:${tokenModelId}`, model_identifier: modelIdentifier })
  //                 .then((response) => {
  //                   // console.log('Deployment status response:', response);
  //                   if (response.data.model_id && response.data.deploy_status === 'complete') {
                      
  //                     // console.log('The NER model is already deployed, and deployment ID is: ', response.data.model_id)
  //                     setDeployStatus('Deployed')
                      
  //                     // Now, list deployments using the deployment_id from the response
  //                     listDeployments(response.data.model_id)
  //                     .then((deployments) => {
  //                       console.log(deployments);
  //                       if (deployments.length > 0) {
  //                           const firstDeployment = deployments[0];
  //                           setNerRAGEndpoint(firstDeployment.modelID);
  //                       }
  //                     })
  //                     .catch((error) => {
  //                         console.error('Error listing deployments:', error);
  //                     });

  //                   } else if (response.data.deploy_status === 'in_progress') {
      
  //                     // console.log('The NER model is still deploying')
  //                     setDeployStatus('Deploying')

  //                   } else {
      
  //                     // console.log('The NER model is not yet deployed and ready to deploy')
  //                     setDeployStatus('Ready to Deploy')

  //                   }
  //                 })
  //                 .catch((error) => {
  //                   if (error.response && error.response.status === 400) {
  //                     // console.log('The NER model is not yet deployed and ready to deploy');
  //                     setDeployStatus('Ready to Deploy')
  //                   } else {
  //                     console.error('Error fetching deployment status:', error);
  //                   }
  //                 });
  //               }

  //             } else if (response.data.deploy_status === 'in_progress') {

  //               // console.log('The NDB model is still deploying')
  //               setDeployStatus('Deploying')

  //             } else {

  //               // console.log('The NDB model is not yet deployed and ready to deploy')
  //               setDeployStatus('Ready to Deploy')
  //             }
  //           })
  //           .catch((error) => {
  //             if (error.response && error.response.status === 400) {
  //               // console.log('The NDB model is not yet deployed and ready to deploy');
  //               setDeployStatus('Ready to Deploy')
  //             } else {
  //               console.error('Error fetching deployment status:', error);
  //             }
  //           });
  //         }
  //       } else if (model.type === 'ndb' || model.type === 'udt') {
        
  //         getDeployStatus({ deployment_identifier: deploymentIdentifier, model_identifier: modelIdentifier })
  //           .then((response) => {
  //             console.log('Deployment status response:', response);
  //             if (response.data.model_id && response.data.deploy_status === 'complete') {
  //               setDeployStatus('Deployed');
  //               setDeploymentId(response.data.model_id);
  //             } else if (response.data.deploy_status === 'in_progress') {
  //               setDeployStatus('Deploying');
  //             } else {
  //               setDeployStatus('Ready to Deploy');
  //             }
  //           })
  //           .catch((error) => {
  //             if (error.response && error.response.status === 400) {
  //               // console.log('Model is not deployed.');
  //               setDeployStatus('Ready to Deploy');
  //             } else {
  //               console.error('Error fetching deployment status:', error);
  //             }
  //           });
  //         }
  //     };

  //     fetchDeployStatus(); // Initial fetch

  //     const intervalId = setInterval(fetchDeployStatus, 2000); // Fetch every 2 seconds

  //     // Cleanup interval on component unmount
  //     return () => clearInterval(intervalId);
  //   }
  // }, [deploymentIdentifier, model]);

  function goToEndpoint() {
    switch (workflow.type) {
      case "semantic_search": {
        let ifGenerationOn = false; // false if semantic search, true if RAG
        let ifGuardRailOn = false; // enable based on actual config
        let guardRailEndpoint = '...'; // change based on actual config
        const newUrl = `/semantic-search/${workflow.id}?ifGenerationOn=${ifGenerationOn}&ifGuardRailOn=${ifGuardRailOn}&guardRailEndpoint=${guardRailEndpoint}`;
        window.open(newUrl, '_blank');
        break;
      }
      case "nlp": {
        router.push(`/token-classification/${workflow.id}`);
        break;
      }
      case "rag": {
        // console.log('workflow.use_llm_guardrail', workflow.use_llm_guardrail);
        // console.log('nerRAGEndpoint', nerRAGEndpoint);
  
        // if (workflow.use_llm_guardrail && nerRAGEndpoint) {
        //   let ifGenerationOn = true; // true for RAG
        //   let ifGuardRailOn = true; // enable based on actual config
        //   let guardRailEndpoint = nerRAGEndpoint; // change based on actual config
        //   const newUrl = `/semantic-search/${deploymentId}?ifGenerationOn=${ifGenerationOn}&ifGuardRailOn=${ifGuardRailOn}&guardRailEndpoint=${guardRailEndpoint}`;
        //   window.open(newUrl, '_blank');
        // } else {
        //   let ifGenerationOn = true; // true for RAG
        //   let ifGuardRailOn = false; // enable based on actual config
        //   let guardRailEndpoint = '...'; // change based on actual config
        //   const newUrl = `/semantic-search/${deploymentId}?ifGenerationOn=${ifGenerationOn}&ifGuardRailOn=${ifGuardRailOn}&guardRailEndpoint=${guardRailEndpoint}`;
        //   window.open(newUrl, '_blank');
        // }
        // break;
      }
      default:
        throw new Error(`Invalid workflow type ${workflow.type}`);
        break;
    }
  }

  // const checkAndDeployNERModel = (tokenModelId: string | undefined) => {

  //     // Check and deploy NER model
  //     if (! tokenModelId) {
  //       setDeployStatus('Deployed')
  //     } else {
  //       const [tokenUsername, modelName] = tokenModelId.split('/');

  //       getDeployStatus({ deployment_identifier: `${tokenModelId}:${tokenModelId}`, model_identifier: modelIdentifier })
  //       .then((response) => {
  //         console.log('Deployment status response:', response);
  //         if (response.data.model_id && response.data.deploy_status === 'complete') {
            
  //           console.log('The NER model is already deployed, and deployment ID is: ', response.data.model_id)
  //           setDeployStatus('Deployed')

  //           // Now, list deployments using the deployment_id from the response
  //           listDeployments(response.data.model_id)
  //           .then((deployments) => {
  //             console.log(deployments);
  //             if (deployments.length > 0) {
  //                 const firstDeployment = deployments[0];
  //                 setNerRAGEndpoint(firstDeployment.modelID);
  //             }
  //           })
  //           .catch((error) => {
  //               console.error('Error listing deployments:', error);
  //           });

  //         } else if (response.data.deploy_status === 'in_progress') {

  //           console.log('The NER model is still deploying')
  //           setDeployStatus('Deploying')

  //         } else {

  //           console.log('The NER model is not yet deployed and ready to deploy')
  //           setDeployStatus('Ready to Deploy')

  //           deployModel({ deployment_name: modelName, model_identifier: tokenModelId })
  //             .then((response) => {
  //               if(response.status === 'complete') {
  //                 console.log('deployment success')

  //                 setDeployStatus('Deployed')
  //               }

  //             })
  //             .catch((error) => {
  //               console.error('Error deploying model:', error);
  //             });

  //         }
  //       })
  //       .catch((error) => {
  //         if (error.response && error.response.status === 400) {
  //           console.log('The NER model is not yet deployed and ready to deploy');
  //           setDeployStatus('Deploying')

  //           deployModel({ deployment_name: modelName, model_identifier: tokenModelId })
  //             .then((response) => {
  //               if(response.status === 'complete') {
  //                 console.log('deployment success')

  //                 setDeployStatus('Deployed')
  //               }

  //             })
  //             .catch((error) => {
  //               console.error('Error deploying model:', error);
  //             });
  //         } else {
  //           console.error('Error fetching deployment status:', error);
  //         }
  //       });
  //     }
  // }

  const [isValid, setIsValid] = useState(false);

  useEffect(() => {
    const validateInterval = setInterval(async () => {
      try {
        const validationResponse = await validate_workflow(workflow.id);
        if (validationResponse.status == 'success') {
          setIsValid(true);
          console.log('Validation passed.');
        } else {
          setIsValid(false);
          console.log('Validation failed.');
        }
      } catch (e) {
        setIsValid(false);
        console.error('Validation failed.', e);
      }
    }, 3000); // Adjust the interval as needed, e.g., every 5 seconds

    return () => clearInterval(validateInterval);
  }, [workflow.id]);

  const handleDeploy = async () => {
    try {
      if (isValid) {
        await start_workflow(workflow.id);
      } else {
        alert('Cannot deploy. The workflow is not valid.');
      }
    } catch (e) {
      console.error('Failed to start workflow.', e);
      alert('Failed to start the workflow.');
    }
  };

  useEffect(()=>{
    if (workflow.status === 'not_started') {
      setDeployStatus('Ready to Deploy')
    } else if (workflow.status === 'in_progress') {
      setDeployStatus('Deploying')
    } else {
      setDeployStatus('Deployed')
    }
  },[workflow.status])

  useEffect(()=>{
    if (workflow.type === 'semantic_search') {
      setDeployType('Semantic Search')
    } else if (workflow.type === 'nlp') {
      setDeployType('Natural Language Processing')
    } else if (workflow.type === 'rag') {
      setDeployType('Retrieval Augmented Generation')
    }
  },[workflow.type])


  return (
    <TableRow>
      <TableCell className="hidden sm:table-cell">
        <Image
          alt="workflow image"
          className="aspect-square rounded-md object-cover"
          height="64"
          src={'/thirdai-small.png'}
          width="64"
        />
      </TableCell>
      <TableCell className="font-medium">{workflow.name}</TableCell>
      <TableCell>
        <Badge variant="outline" className="capitalize">
          {deployStatus}
        </Badge>
      </TableCell>
      <TableCell className="hidden md:table-cell">{deployType}</TableCell>
      <TableCell className="hidden md:table-cell">
        {
          new Date(workflow.publish_date).toLocaleDateString('en-US', {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
            })
        }
      </TableCell>
      <TableCell className="hidden md:table-cell">'N\A'</TableCell>
      <TableCell className="hidden md:table-cell">
        <button type="button" 
                onClick={goToEndpoint}
                className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-full text-sm p-2.5 text-center inline-flex items-center me-2 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">
          <svg className="w-4 h-4" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 10">
          <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M1 5h12m0 0L9 1m4 4L9 9"/>
          </svg>
          <span className="sr-only">Go to endpoint</span>
        </button>
      </TableCell>
      <TableCell className="hidden md:table-cell">
        {
          isValid
          &&
          <button type="button" 
                  onClick={handleDeploy}
                  className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-full text-sm p-2.5 text-center inline-flex items-center me-2 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">
            <svg className="w-4 h-4" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 10">
              <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M1 5h12m0 0L9 1m4 4L9 9"/>
            </svg>
            <span className="sr-only">Deploy</span>
          </button>
        }
      </TableCell>
      <TableCell>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button aria-haspopup="true" size="icon" variant="ghost">
              <MoreHorizontal className="h-4 w-4" />
              <span className="sr-only">Toggle menu</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            <DropdownMenuItem>Edit</DropdownMenuItem>
            {
              deployStatus === 'Deployed' && deploymentIdentifier
              &&
              <>
              <DropdownMenuItem>
                <form action={deleteModel}>
                  <button type="button"
                  onClick={()=>{console.log('undeploy workflow')}}
                  >Undeploy</button>
                </form>
              </DropdownMenuItem>

              <Link href={`/analytics?id=${encodeURIComponent(`${workflow.id}`)}`}>
                <DropdownMenuItem>
                    <button type="button">Usage stats</button>
                </DropdownMenuItem>
              </Link>
              </>
            }
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  );
}
