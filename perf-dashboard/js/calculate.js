// reference: https://blog.harveydelaney.com/parsing-a-csv-file-using-node-javascript/
var fs = require("fs");
var parse = require("csv-parse");

var csvFile = "../data/data.csv";

class User {
    constructor(id, firstName, lastName, age, email, gender, country) {
        this.id = id;
        this.firstName = firstName;
        this.lastName = lastName;
        this.age = age;
        this.email = email;
        this.gender = gender;
        this.country = country;
    }
}

const processData = (err, data) => {
    if (err) {
        console.log(`An error was encountered: ${err}`);
        return;
    }

    data.shift(); // only required if csv has heading row

    const userList = data.map(row => new User(...row));

    analyseUsers(userList);
}

fs.createReadStream(csvFile)
    .pipe(parse({ delimiter: ',' }, processData));

const analyseUsers = userList => {
    const ageSum = userList.reduce((acc, val) => acc += val.age, 0);
    const averageAge = ageSum / userList.length;
    console.log(averageAge);
}

