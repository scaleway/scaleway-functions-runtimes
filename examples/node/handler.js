exports.handle = async (event, context) => {
    return {
        body: {
            message: 'Hello World',
        },
        statusCode: 201
    }
};
