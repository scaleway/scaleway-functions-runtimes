import json

def handle(event, context):
    return {
        'body': json.dumps({'test': 'test', 'tutu': 12}),
        'statusCode': 201
    }
