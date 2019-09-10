import * as express from 'express';
import * as bodyParser from 'body-parser';

let handler: any;

// ---- Configuration ---- //

const app = express();
app.use(bodyParser.json());
app.use(bodyParser.raw());
app.use(bodyParser.text({ type : "text/*" }));
app.disable('x-powered-by');

// -- Server Logic -- //

/**
 * Handle HTTP Response from customer's function handler results (after execution)
 * @param {express.Response} res - Response Object to handle HTTP Response management
 * @param {Object|string} functionResult - Result retrieved from customer => Function handler's execution
 */
const handleResponse = (res: express.Response, functionResult: any) => {
    return res.status(200).send(functionResult);
};


/**
 * This is the function's Gateway, it is in charge of managing incoming HTTP traffic, transform incoming requests
 * into usable events, assemble the execution context, execute the handler with formatted event and context
 * and handle the response by formatting the Handler's output into an HTTP response.
 * @param req {Request} Express HTTP Request object
 * @param res {Response} Express HTTP Response object
 * @returns {Promise<void>}
 */
const functionGateway = async (req: express.Request, res: express.Response) => {
    let responseSent = false;
    const callback = (err: Error, functionResult: any) => {
        responseSent = true;
        if (err) {
            console.error(err);
            return res.status(500).send(err);
        }
        return handleResponse(res, functionResult);
    };

    if (!handler) {
        const handlerName = req.body.handlerName;
        const handlerFilePath = req.body.handlerPath;
        try {
            handler = require(handlerFilePath)[handlerName];
        } catch (e) {
            return res.status(500).send('Function Handler does not exist, check that you provided the right HANDLER parameter (path to your module with exported function to use)')
        }
    }

    // When building JavaScript from TypeScript, exported members from a module are under the format: { default: [Function (Handler)] }
    // immport dynamically
    if (typeof handler !== 'function') {
        return res.status(500).send('Provided Handler does not exist, or does not export methods properly.');
    }

    try {
        const functionResult = await handler(req.body.event, req.body.context, callback);
        // Response has been sent via Callback
        if (responseSent || !functionResult) return;
        return handleResponse(res, functionResult);
    } catch (err) {
        return res.status(500).send(err.message);
    }
};

app.all('/*', functionGateway);


const port = process.env.SCW_UPSTREAM_PORT || 8081;

app.listen(port, () => {
    console.log(`Scaleway Node.js listening on port: ${port}`)
})
