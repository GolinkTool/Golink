import os
import subprocess
import time
import requests
import re
import argparse
import json
from packaging.version import Version, InvalidVersion
'''
This script is used to obtain the correct version of the target project by using the binary search method.
The method doesn't specify the error dependency, instead it extracts the target error dependency from the error log.
Then go through the tags of the target project to find the correct version.
'''

def execute_command(name, *args, timeout=1200):
    start = time.time()
    cmd = [name] + list(args)

    try:
        result = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=timeout, text=True)
        duration = time.time() - start
        log = f"{result.stdout}\n{result.stderr}"
        return log, duration

    except subprocess.TimeoutExpired as e:
        duration = time.time() - start
        log = f"command timeout: {e.cmd}\ntime: {timeout}s\n"
        return log, duration


def remove_go_mod(project_path):
    go_mod_path = os.path.join(project_path, "go.mod")
    go_sum_path = os.path.join(project_path, "go.sum")

    if os.path.exists(go_mod_path):
        try:
            os.remove(go_mod_path)
        except Exception as e:
            print(f"delete {go_mod_path} Error: {e}")

    if os.path.exists(go_sum_path):
        try:
            os.remove(go_sum_path)
        except Exception as e:
            print(f"delete {go_sum_path} Error: {e}")

def read_file_by_line(file_path):
    with open(file_path, 'r') as file:
        for line in file:
            yield line.strip()


def is_valid_version(version):
    try:
        Version(version)
        return True
    except InvalidVersion:
        return False

def sort_versions(versions):
    valid_versions = [v for v in versions if is_valid_version(v[1:])]

    parsed_versions = [Version(v[1:]) for v in valid_versions]

    sorted_versions = sorted(parsed_versions, reverse=True)

    sorted_versions_str = [f"v{v}" for v in sorted_versions]

    return sorted_versions_str

def read_tpl_tags(target_tpl_tag_path):
    try:
        with open(target_tpl_tag_path, 'r') as file:
            lines = file.readlines()
    except Exception as e:
        print(f"An error occurred while reading the tag file: {e}")
        return []

    versions = [line.split('@@')[1].strip() for line in lines]
    sorted_versions = sort_versions(versions)

    return sorted_versions


def append_to_file(file_path, data):
    try:
        with open(file_path, 'a') as file:
            file.write(data + '\n')
    except Exception as e:
        print(f"An error occurred while writing to the file: {e}")


def check_module_file(url):

    try:
        response = requests.head(url)
        if response.status_code == 200:
            return True
        else:
            return False
    except Exception as e:
        return False


def extract_module_path(url):
    try:
        response = requests.get(url)
        if response.status_code == 200:
            content = response.text
            match = re.search(r'module\s+([^\s"]+)', content)
            if match:
                module_path = match.group(1)
                return module_path
            else:
                return ""
    except requests.RequestException as e:
        return ""

def extract_error_reason(log):
    '''
        MAIN ERRORS :
        Module Path Mismatch: @latest found (versionxxx), but does not contain package; undefined
        API Miss: but was required as
    '''
    reason = ""
    possible_reason = ["undefined", "but does not contain package","@latest found","but was required as",
                       "proxyconnect tcp: EOF","TLS handshake timeout",
                       "checksum mismatch",
                       "ambiguous import",
                       "cannot find module providing package","can't request version",
                       "relative import paths are not supported in module mode",
                       "case-insensitive file name collision",
                       "no Go files in", "unknown revision",
                       "but not marked as explicit in vendor","No such file or directory",
                       "connection reset by peer","Invalid username or password",
                       "matched no packages",
                       "no matching versions for query","invalid version","invalid char",
                       " can't request explicit version","Internal Server Error",
                       "should be v0 or v1, not v2","used for two different module paths",
                       "invalid: malformed module path","Error in the HTTP2 framing layer"
                       ]
    for r in possible_reason:
        if log.__contains__(r):
            reason += r + ","

    return reason[:-1]

def AppendAPI(log, replace_API, log_file_name):
    error_reason = extract_error_reason(log)
    print(f"Error Reason: {error_reason}")
    if error_reason == "":
        return False
    if "but was required as" in log:
        module_declares_match = re.search(r'module declares its path as:\s*(.*)', log)
        required_as_match = re.search(r'but was required as:\s*(.*)', log)
        if module_declares_match and required_as_match:
            declared_path = module_declares_match.group(1)
            required_path = required_as_match.group(1)
            print(f"Error Path: required {required_path} => declared {declared_path}")
            replace_API[required_path] = declared_path
            append_to_file(os.path.join(LOG_DIR, log_file_name), "\n**replace -> " + required_path +"to " + declared_path)

    return True

def ReplaceFile(replace_API, project_path, project_name, sum_time, version):
    command = f"cd {project_path}&& go mod init {project_name}"
    log, init_time = execute_command("sh", "-c", command)
    sum_time += init_time
    go_mod_path = os.path.join(project_path, "go.mod")
    replace_start = time.time()
    for oldAPI, newAPI in replace_API.items():
        if version.startswith("v1") or version.startswith("v0"):    # to obey the Semantic Versioning Rules(SemVer)
            pass
        else:
            newAPI = newAPI + '/' + version[:2]
        with open(go_mod_path, 'a') as go_mod_file:
            go_mod_file.write(f'\nreplace {oldAPI} => {newAPI} {version}\n')
    replace_time = time.time() - replace_start
    sum_time += replace_time
    return


def bs_approach(project_path,project_name, result_log):
    sum_time = 0
    log_file_name = project_name.replace("/", "@@") + ".log"
    print(project_path, project_name, log_file_name)
    # constants for loop
    iteration_times = 0
    build_success = False
    BS_result = ""

    replace_API = {}    # replace all the APIs that have met before
    remove_go_mod(project_path)

    command = f"cd {project_path} && go mod init {project_name} && go mod tidy"
    log, go_build_time = execute_command("sh", "-c", command)
    print(log)
    sum_time += go_build_time
    iteration_times += 1
    append_to_file(os.path.join(LOG_DIR, log_file_name), f"\nNo.{iteration_times} round costs {go_build_time}s, total time: {sum_time}s")

    tpl_filename_slice = TARGET_TPL.split('/')[:3]
    tpl_tag_file_name = '@@'.join(tpl_filename_slice) + '.txt'
    tags = read_tpl_tags(os.path.join(TARGET_TAGS_DIR, tpl_tag_file_name))
    # ----------------------------------------BS Begin--------------------------------------------------------------------------------------
    for version in tags:
#         clean_cache()
        if sum_time > 3600:
            append_to_file(os.path.join(LOG_DIR, log_file_name), "Operation timed out; it has been running for 1 hour.")
            BS_result = "Timeout"
            break

        start = time.time()
        remove_go_mod(project_path)
        remove_time = time.time() - start
        append_to_file(os.path.join(LOG_DIR, log_file_name), f"update to version {version} for target TPL {TARGET_TPL}, begin to build")

        url = f"https://{TARGET_TPL}/blob/{version}/go.mod"
        if check_module_file(url):
            module_path = extract_module_path(url)
            if len(replace_API) > 0:
                ReplaceFile(replace_API, project_path, project_name, sum_time, version)
                command = f"cd {project_path} && go get {module_path}@{version} && go mod tidy"
            else:
                command = f"cd {project_path} && go mod init {project_name} && go get {module_path}@{version} && go mod tidy "

        else:
            if len(replace_API) > 0:
                ReplaceFile(replace_API, project_path, project_name, sum_time, version)
                if not (version.startswith("v1") or version.startswith("v0")):
                    command = f"cd {project_path} && go get {TARGET_TPL}/{version[:2]}@{version} && go mod tidy"
                else:
                    command = f"cd {project_path} && go get {TARGET_TPL}@{version} && go mod tidy "
            else:
                command = f"cd {project_path} && go mod init {project_name} && go get {TARGET_TPL}@{version} && go mod tidy "

        print(str(iteration_times+1) + "->" + command)
        log, go_build_time = execute_command("sh", "-c", command)
        sum_time += remove_time
        sum_time += go_build_time
        iteration_times += 1
        # Notice : Here sum_time includes go_build_time, replace_time and clean_time
        append_to_file(os.path.join(LOG_DIR, log_file_name), f"No.{iteration_times} round costs {go_build_time}s, total time: {sum_time}s")

        if not AppendAPI(log, replace_API, log_file_name):
            append_to_file(os.path.join(LOG_DIR, log_file_name), "Successfully obtained the correct version:" + version + "\n\n" + log)
            build_success = True
            break
        elif "but was required as" in log:
            remove_go_mod(project_path)
            ReplaceFile(replace_API, project_path, project_name, sum_time, version)
            if not (version.startswith("v1") or version.startswith("v0")):
                    command = f"cd {project_path} && go get {TARGET_TPL}/{version[:2]}@{version} && go mod tidy"
            else:
                command = f"cd {project_path} && go get {TARGET_TPL}@{version} && go mod tidy "
            print(str(iteration_times+1) + "->" + command)
            log, go_build_time = execute_command("sh", "-c", command)
            sum_time += go_build_time
            iteration_times += 1

            append_to_file(os.path.join(LOG_DIR, log_file_name), f"No.{iteration_times} round costs {go_build_time}s, total time: {sum_time}s")

            error_reason = extract_error_reason(log)
            print(f"Error Reason: {error_reason}")
            if error_reason == "":
                build_success = True
                print("Successfully obtained the correct version, extraction error is empty")
                append_to_file(os.path.join(LOG_DIR, log_file_name), "Successfully obtain the right version:" + TARGET_TPL + "@" + version + "\n\n" + log)
                break
            else:
                append_to_file(os.path.join(LOG_DIR, log_file_name), "Fail: " + json.dumps(replace_API) + '\n' + version + "\n" + error_reason + log)
                continue
        else:
            build_success = False
            append_to_file(os.path.join(LOG_DIR, log_file_name), "Fail: " + json.dumps(replace_API) + '\n' + version + '\n' + log)
            continue

    if not build_success and BS_result == "":
        BS_result = "Fail"

    if build_success:
        append_to_file(os.path.join(LOG_DIR, log_file_name),
                    f"{project_name}@{TARGET_TPL}@{iteration_times}@{sum_time}@{'BS Success'}@{version}")
        print(f"{project_name}@{TARGET_TPL}@{iteration_times}@{sum_time}@{'BS Success'}@{version}")
        append_to_file(os.path.join(LOG_DIR, result_log), f"{project_name}@{TARGET_TPL}@{iteration_times}@{sum_time}@{'BS Success'}@{version}")
    else:
        append_to_file(os.path.join(LOG_DIR, log_file_name),
                   f"{project_name}@{TARGET_TPL}@{iteration_times}@{sum_time}@{BS_result}")
        print(f"{project_name}@{TARGET_TPL}@{iteration_times}@{sum_time}@{BS_result}")
        append_to_file(os.path.join(LOG_DIR, result_log), f"{project_name}@{TARGET_TPL}@{iteration_times}@{sum_time}@{BS_result}")



if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='BS script for Go projects')
    parser.add_argument('--project_name', type=str, required=True, help='Path to the downloaded Go projects directory')
    parser.add_argument('--working_dir', type=str, required=True, help='Working directory')
    parser.add_argument('--target_tags_dir', type=str, required=True, help='Directory to store the tags of the target project')
    parser.add_argument('--log_dir', type=str, required=True, help='Directory to store logs')
    parser.add_argument('--target_tpl', type=str, required=True, help='Third-party library need backtracking search.')

    args = parser.parse_args()
    PROJECT_NAME = args.project_name
    WORKING_DIR = args.working_dir
    TARGET_TAGS_DIR = args.target_tags_dir
    LOG_DIR = args.log_dir
    TARGET_TPL = args.target_tpl
    RESULT_LOG = "BS_approach.log"
    PROJECT_PATH = os.path.join(WORKING_DIR,PROJECT_NAME)

    bs_approach(PROJECT_PATH,PROJECT_NAME, RESULT_LOG)
