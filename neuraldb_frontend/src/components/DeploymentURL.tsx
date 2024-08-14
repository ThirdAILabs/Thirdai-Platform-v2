// ----------------------------------------------------------------------

export function createDeploymentUrl(deploymentId: string) {
    const hostname = `${window.location.protocol}//${window.location.host}`;
    return `${hostname}/${deploymentId}`;
}

export function createTokenModelUrl(deploymentId: string) {
    const hostname = `${window.location.protocol}//${window.location.host}`;
    return `${hostname}/${deploymentId}`;
}
