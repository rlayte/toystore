# toystore

Dynamo implementation

## Setup

We're using [supervisor](http://supervisord.org) to manage multiple process required for testing. Install using pip:

    $ sudo pip install supervisor

Start the cluster:

    $ supervisord

This will start servers on http://localhost:{3000,3001,3002,3003,3004} by default.
