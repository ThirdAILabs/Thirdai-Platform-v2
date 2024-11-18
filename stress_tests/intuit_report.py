ndbimport os
import subprocess


# def run_stress_test(
#     predict_weight, 
#     insert_weight, 
#     delete_weight,
#     initial_config,
#     docs_per_insertion,
#     file_size,
#     dev_vs_prod_mode,
#     max_concurrent_users,
#     spawn_rate,
#     run_time, 
#     min_wait,
#     max_wait,
# ):
#     size_to_docs = {
#         "1MB": "/home/david/intuit-stress-test/pubmed_1000_files_1MB_each", 
#         "10MB": "/home/david/intuit-stress-test/pubmed_100_files_10MB_each", 
#         "100MB": "/home/david/intuit-stress-test/pubmed_10_files_100MB_each"
#     }
#     docs_folder = size_to_docs[file_size]
#     autoscaling_enabled = "--autoscaling_enabled" if dev_vs_prod_mode == "prod" else ""
#     cwd = os.path.dirname(os.path.dirname(__file__))
#     command = (
#         f"python3 -m stress_tests.end_to_end_stress_test --host http://3.91.36.22 --email david@thirdai.com --password password --config {initial_config} --users {max_concurrent_users} --spawn_rate {spawn_rate} --run_time {run_time} {autoscaling_enabled} --docs_folder {docs_folder} --docs_per_insertion {docs_per_insertion} --min_wait {min_wait} --max_wait {max_wait} --predict_weight {predict_weight} --insert_weight {insert_weight} --delete_weight {delete_weight}",
#     )
#     result = subprocess.run(command, shell=True, cwd=cwd)


def run_stress_test(
    predict_weight, 
    insert_weight, 
    delete_weight,
    initial_config,
    docs_per_insertion,
    file_size,
    dev_vs_prod_mode,
    max_concurrent_users,
    spawn_rate,
    run_time, 
    min_wait,
    max_wait,
):
    size_to_docs = {
        "1MB": "/home/david/intuit-stress-test/pubmed_1000_files_1MB_each", 
        "10MB": "/home/david/intuit-stress-test/pubmed_100_files_10MB_each", 
        "100MB": "/home/david/intuit-stress-test/pubmed_10_files_100MB_each"
    }
    docs_folder = size_to_docs[file_size]
    autoscaling_enabled = "--autoscaling_enabled" if dev_vs_prod_mode == "prod" else ""
    folder = os.path.dirname(__file__)
    script_path = os.path.join(folder, "stress_test_deployment.py")
    command = (
        f"locust -f {script_path} --headless --host http://3.91.36.22 --email david@thirdai.com --password password --deployment_id 7dee933b-23ea-4da8-a076-3cd2a1496731 --users {max_concurrent_users} --spawn-rate {spawn_rate} --run-time {run_time} --docs_folder {docs_folder} --docs_per_insertion {docs_per_insertion} --min_wait {min_wait} --max_wait {max_wait} --predict_weight {predict_weight} --insert_weight {insert_weight} --delete_weight {delete_weight}",
        # f"locust -f {script_path} --headless --host https://cerulean-stable.thirdai-aws.com --email admin@thirdai.com --password butSenator-IamSingaporean! --deployment_id 4fab105e-13a0-4445-8f98-0abecf6feba7 --users {max_concurrent_users} --spawn-rate {spawn_rate} --run-time {run_time} --docs_folder {docs_folder} --docs_per_insertion {docs_per_insertion} --min_wait {min_wait} --max_wait {max_wait} --predict_weight {predict_weight} --insert_weight {insert_weight} --delete_weight {delete_weight}",
    )
    result = subprocess.run(command, shell=True)
    

run_stress_test(
    predict_weight=18, 
    insert_weight=1, 
    delete_weight=0,
    initial_config="small-pdf",
    docs_per_insertion=40,
    file_size="1MB",
    dev_vs_prod_mode="dev",
    max_concurrent_users=1,
    spawn_rate=1,
    run_time=200, 
    min_wait=3,
    max_wait=5,
)




# 20% tests are with 18 2 1
# 40% tests are with 18 4 2

# first set of experiments:
# run_stress_test(
#     predict_weight=18, 
#     insert_weight=2, 
#     delete_weight=1,
#     initial_config="small-pdf",
#     docs_per_insertion=1,
#     file_size="1MB",
#     dev_vs_prod_mode="dev",
#     max_concurrent_users=10,
#     spawn_rate=10,
#     run_time=200, 
#     min_wait=10,
#     max_wait=20,
# )


# prod mode sanity check
# run_stress_test(
#     predict_weight=18, 
#     insert_weight=4, 
#     delete_weight=2,
#     initial_config="small-pdf",
#     docs_per_insertion=1,
#     file_size="1MB",
#     dev_vs_prod_mode="dev",
#     max_concurrent_users=200,
#     spawn_rate=10,
#     run_time=200, 
#     min_wait=2,
#     max_wait=5,
# )