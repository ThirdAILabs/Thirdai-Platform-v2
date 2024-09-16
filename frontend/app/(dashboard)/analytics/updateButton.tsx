'use client';

import { Button } from "@/components/ui/button";
import { retrain_ndb } from "@/lib/backend"; // Ensure the correct path
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";

export default function UpdateButton() {
    const params = useSearchParams();
    const router = useRouter();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const workflowId = params.get("id");
    const username = params.get("username");
    const model_name = params.get("model_name");

    console.log('workflowId', 'username', 'model_name', workflowId, username, model_name)
    console.log('base_model_identifier', `${username}/${model_name}`)

    /**
     * Handles the update button click by initiating the retrain process.
     */
    async function handleUpdate() {
        const workflowId = params.get("id");
        const username = params.get("username");
        const model_name = params.get("model_name");

        // Validate required parameters
        if (!workflowId || !username || !model_name) {
            setError("Missing required parameters: id, username, or model_name.");
            return;
        }

        // Define the base_model_identifier in the format 'username/model_name'
        const base_model_identifier = `${username}/${model_name}`;

        // Define job options as per your requirements
        const job_options = {
            allocation_cores: 4, // Example value
            allocation_memory: 8192, // Example value in MB
            // Add other JobOptions fields as necessary
        };

        setLoading(true);
        setError(null);

        try {
            const data = await retrain_ndb({ model_name, base_model_identifier, job_options });
            console.log("Retrain initiated successfully:", data);
            // TODO: navigate or update the UI based on the response
            // router.push(`/analytics?id=${encodeURIComponent(`${workflowId}-updated`)}&username=${encodeURIComponent(username)}&model_name=${encodeURIComponent(model_name)}`);
        } catch (err: any) {
            console.error("Error retraining model:", err);
            setError(err.message);
            alert("Error retraining model: " + err.message);
        } finally {
            setLoading(false);
        }
    }
    
    return (
        <div style={{ display: "flex", justifyContent: "center", marginTop: "20px", marginBottom: "20vh" }}>
            <Button onClick={handleUpdate} disabled={loading}>
                {loading ? "Updating..." : "Update model with feedback"}
            </Button>
            {error && <p style={{ color: "red", marginTop: "10px" }}>{error}</p>}
        </div>
    );
}