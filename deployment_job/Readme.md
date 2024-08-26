1. **Configure Environment:**
   - Open the `.env` file in `deployment_job` folder 
   - Change `LICENSE_KEY` with your thirdai license key
   - Do not change the `CHECKPOINT_DIR` variable

2. **Create docker image:**
   - Build docker image with the command
   ```
   docker build -t <image_name>:<tag> .
   ```

3. **Run docker container:**  
   - Go to the `deployment_job` folder 
   - In this command, replace 
      -  `local_port` with any available port on your local machine
      -  `checkpoint_dir`with your model checkpoint directory, containing `model.udt`
      -  `env_file_path` path to your env file (in step 1)
   ```
   docker run -p <local_port>:80 --env-file <env_file_path> -v <checkpoint_dir>:/pretrained_model <image_name>:<tag>
   ```

4. **Predict endpoint:**
   - The container will be running on `localhost:7888` (ex. local_port is 7888)
   - Run the following curl command to make a predict call
   ```
   curl -X POST \
  'http://localhost:7889/predict' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "What is artificial intelligence?",
    "top_k": 1
  }'
   ```