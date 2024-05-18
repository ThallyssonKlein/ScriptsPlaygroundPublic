import notion from "../index.js";

notion.databases.query({
    database_id: process.env.TASKS_DATABASE_ID,
    filter: {
        property: "Source",
        select: {
            equals: "Rotina"
        }
    }

}).then(response => {
    console.log(JSON.stringify(response));
})