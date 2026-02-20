import {
    ExtractPermsList,
    PermedUser,
    UserHasPerm,
} from "@/api/permissions/permshelpers";

describe("Perms Helpers", () => {
    it("ExtractPermsList should extract perms correctly", () => {
        const user: PermedUser = {
            user_id: 5,
            permissions: [
                {
                    description: "Lets you read users",
                    permission: "read:user",
                    name: "Read Users",
                },
                {
                    description: "Lets you write users",
                    permission: "write:user",
                    name: "Write Users",
                },
                {
                    description: "Lets you read images",
                    permission: "read:images",
                    name: "Read Images",
                },
            ],
        };

        const result = ExtractPermsList(user);
        expect(result).toStrictEqual([
            "read:user",
            "write:user",
            "read:images",
        ]);
    });

    it("ExtractPermsList should handle null permissions well", () => {
        const user: PermedUser = { user_id: 67 };

        expect(ExtractPermsList(user)).toBeUndefined();
    });

    it("UserHasPerm should properly approve existing perm", () => {
        const user: PermedUser = {
            user_id: 5,
            permissions: [
                {
                    description: "Lets you read users",
                    permission: "read:user",
                    name: "Read Users",
                },
                {
                    description: "Lets you write users",
                    permission: "write:user",
                    name: "Write Users",
                },
                {
                    description: "Lets you read images",
                    permission: "read:images",
                    name: "Read Images",
                },
            ],
        };

        expect(UserHasPerm(user, "read:user")).toBeTruthy();
    });

    it("UserHasPerm should properly disapprove nonexisting perm", () => {
        const user: PermedUser = {
            user_id: 5,
            permissions: [
                {
                    description: "Lets you read users",
                    permission: "read:user",
                    name: "Read Users",
                },
                {
                    description: "Lets you write users",
                    permission: "write:user",
                    name: "Write Users",
                },
                {
                    description: "Lets you read images",
                    permission: "read:images",
                    name: "Read Images",
                },
            ],
        };

        expect(UserHasPerm(user, "write:images")).toBeFalsy();
    });

    it("UserHasPerm should properly disapprove on undefined perms", () => {
        const user: PermedUser = { user_id: 67 };

        expect(UserHasPerm(user, "write:images")).toBeFalsy();
    });
});
