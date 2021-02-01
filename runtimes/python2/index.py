from flask import Flask, request, jsonify
from importlib import import_module
import json
import sys

def import_function_handler(file_path, handler_name):
    split_module_path = file_path.split('/')

    module_index = split_module_path.__len__() - 1
    module_name = split_module_path[module_index]

    package_name = '/'.join(split_module_path[:module_index])

    sys.path.append(package_name)
    try:
        module = import_module(module_name)
        return getattr(module, handler_name)
    except Exception as e:
        print(e)
        # Raise exception with custom error message for UX
        raise Exception('Function Handler does not exist, check that you provided the right HANDLER parameter (path to your module with exported function to use), check your function logs')

app = Flask(__name__)

@app.route("/", defaults={"path": ""}, methods=["POST"])
@app.route("/<path:path>", methods=["POST"])
def main_route(path):
    # Dynamically import module
    body = json.loads(request.get_data())
    try:
        function_handler = import_function_handler(body.get('handlerPath'), body.get('handlerName'))
        function_result = function_handler(body.get('event'), body.get('context'))
    except Exception as e:
        return str(e), 500

    # If function response is already a string/json encoded -> send HTTP response
    if isinstance(function_result, basestring):
        return function_result
    return jsonify(function_result), 200
