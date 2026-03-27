"use client";

import { createContext, useCallback, useState } from "react";

export type ProfilePictureContextValue = {
    /** Increments every time the current user's profile picture is updated. */
    profilePictureVersion: number;
    /** Call this after a successful profile picture upload to notify all listeners. */
    refreshProfilePicture: () => void;
};

export const ProfilePictureContext =
    createContext<ProfilePictureContextValue>({
        profilePictureVersion: 0,
        refreshProfilePicture: () => {},
    });

/**
 * Provides a lightweight version counter so that any component that displays
 * the current user's profile picture (e.g. the Sidebar) can re-fetch it
 * whenever the Settings page successfully uploads a new one.
 */
export function ProfilePictureProvider({
    children,
}: {
    children: React.ReactNode;
}) {
    const [profilePictureVersion, setProfilePictureVersion] = useState(0);

    const refreshProfilePicture = useCallback(() => {
        setProfilePictureVersion((v) => v + 1);
    }, []);

    return (
        <ProfilePictureContext
            value={{ profilePictureVersion, refreshProfilePicture }}
        >
            {children}
        </ProfilePictureContext>
    );
}
