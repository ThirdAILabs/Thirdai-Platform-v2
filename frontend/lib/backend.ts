import axios from 'axios';

// /lib/backend.js
export async function fetchPublicModels(name: string) {
    const response = await fetch(`http://localhost:8000/api/model/public-list?name=${name}`);
    if (!response.ok) {
        throw new Error('Failed to fetch public models');
    }
    return response.json();
}

interface TrainNdbParams {
    name: string;
    formData: FormData;
}

export function train_ndb({ name, formData }: TrainNdbParams): Promise<any> {
    return new Promise((resolve, reject) => {
        axios
            .post(`http://localhost:8000/api/train/ndb?model_name=${name}&sharded=${false}`, formData)
            .then((res) => {
                resolve(res.data);
            })
            .catch((err) => {
                if (err.response && err.response.data) {
                    reject(new Error(err.response.data.detail || 'Failed to run model'));
                } else {
                    reject(new Error('Failed to run model'));
                }
            });
    });
}