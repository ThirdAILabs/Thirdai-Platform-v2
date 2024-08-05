// ----------------------------------------------------------------------

export function createDeploymentUrl(deploymentId: string) {
    const hostname = process.env.DEPLOYMENT_BASE_URL;
    return `${hostname}/${deploymentId}`;
}
