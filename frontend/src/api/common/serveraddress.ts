"use server";

export default async function SERVER_ADDRESS() {
    return `http://${process.env.FREEZETAG_BACKEND_ADDRESS}:3824/`;
}
