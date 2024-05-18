import dotenv from 'dotenv'
dotenv.config();

import { Client }  from "@notionhq/client";

export default new Client({
    auth: process.env.NOTION_TOKEN,
});