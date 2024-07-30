// /lib/backend.js

import axios from 'axios';

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
    const access_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOTM4MGQ2NmQtMTIzZi00ODhkLTgwNTUtOTEwNGI5YzBhM2ViIiwiZXhwIjoxNzIyMzU0Mzk1fQ.Tl13wtCHnXes7nPxiI2KkHE4bJYqyXijNmOpjYDRQbA";

    axios.defaults.headers.common.Authorization = `Bearer ${access_token}`;

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

export function userEmailLogin(email: string, password: string): Promise<any> {
    return new Promise((resolve, reject) => {
      axios
        .get('http://localhost:8000/api/user/email-login', {
          headers: {
            Authorization: `Basic ${window.btoa(`${email}:${password}`)}`,
          },
        })
        .then((res) => {
          resolve(res.data);
        })
        .catch((err) => {
          reject(err);
        });
    });
}

export function userRegister(email: string, password: string, username: string) {
    return new Promise((resolve, reject) => {
      axios
        .post('http://localhost:8000/api/user/email-signup-basic', {
          email,
          password,
          username,
        })
        .then((res) => {
          resolve(res.data);
        })
        .catch((err) => {
          reject(err);
        });
    });
}