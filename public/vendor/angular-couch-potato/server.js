var express = require('express');
var app = express();

app.use('/angular-couch-potato', express.static(__dirname + '/dist'));

app.listen(3000); //the port you want to use
