export type Perm = { description: string; name: string; permission: string };
export type PermedUser = {
    user_id: number;
    permissions?: Perm[];
};

export function ExtractPermsList(user: PermedUser): string[] | undefined {
    return user.permissions?.map((v) => v.permission);
}

export function UserHasPerm(user: PermedUser, permission: string): boolean {
    return ExtractPermsList(user)?.includes(permission) ?? false;
}
