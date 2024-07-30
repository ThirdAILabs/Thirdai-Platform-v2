// /lib/backend.js

import axios from 'axios';

function getAccessToken(): string {
  const accessToken = localStorage.getItem('accessToken');
  if (!accessToken) {
    throw new Error('Access token is not available');
  }
  return accessToken;
}

export async function fetchPrivateModels(name: string) {
  // Retrieve the access token from local storage
  const accessToken = getAccessToken()

  // Set the default authorization header for axios
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  try {
    const response = await axios.get(`http://localhost:8000/api/model/list`, {
      params: { name },
    });
    return response.data;
  } catch (error) {
    console.error('Error fetching private models:', error);
    throw new Error('Failed to fetch private models');
  }
}

export async function fetchPublicModels(name: string) {
    const response = await fetch(`http://localhost:8000/api/model/public-list?name=${name}`);
    if (!response.ok) {
        throw new Error('Failed to fetch public models');
    }
    return response.json();
}

export async function fetchPendingModel() {
  // Retrieve the access token from local storage
  const accessToken = getAccessToken()

  // Set the default authorization header for axios
  axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

  try {
    const response = await axios.get(`http://localhost:8000/api/model/pending-train-models`);
    return response.data;
  } catch (error) {
    console.error('Error fetching private models:', error);
    throw new Error('Failed to fetch private models');
  }
}

interface TrainNdbParams {
    name: string;
    formData: FormData;
}

export function train_ndb({ name, formData }: TrainNdbParams): Promise<any> {
    // Retrieve the access token from local storage
    const accessToken = getAccessToken()

    // Set the default authorization header for axios
    axios.defaults.headers.common.Authorization = `Bearer ${accessToken}`;

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
          const accessToken = res.data.data.access_token;

          if (accessToken) {
            // Store accessToken into local storage, replacing any existing one.
            localStorage.setItem('accessToken', accessToken);
          }
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