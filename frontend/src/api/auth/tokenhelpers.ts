"use client";

export const SaveToken = (token: string) => {
    localStorage.setItem("freezetag_token", token);
};

export const GetToken = () => {
    return localStorage.getItem("freezetag_token");
};

export const ClearToken = () => {
    localStorage.removeItem("freezetag_token");
};