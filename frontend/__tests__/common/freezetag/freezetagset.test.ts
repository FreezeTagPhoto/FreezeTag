import FreezeTagSet from "@/common/freezetag/freezetagset";

describe("FreezeTagSet", () => {
    it("should be immutable on add", () => {
        const set1 = new FreezeTagSet([1, 2, 3, 4]);
        const set2 = set1.add(5, 6, 7);

        expect(set1.has(1, 2, 3, 4)).toBeTruthy();
        expect(set1.has(5)).toBeFalsy();
        expect(set1.has(6)).toBeFalsy();
        expect(set1.has(7)).toBeFalsy();
        expect(set2.has(1, 2, 3, 4, 5, 6, 7)).toBeTruthy();
    });

    it("should be immutable on remove", () => {
        const set1 = new FreezeTagSet([1, 2, 3, 4, 5, 6, 7]);
        const set2 = set1.remove(5, 6, 7);

        expect(set1.has(1, 2, 3, 4, 5, 6, 7)).toBeTruthy();
        expect(set2.has(1, 2, 3, 4)).toBeTruthy();
        expect(set2.has(5, 6, 7)).toBeFalsy();
    });

    it("should work with toggle", () => {
        const set1 = new FreezeTagSet([1, 2, 3, 4, 5, 6, 7]);
        const set2 = set1.toggle(7);
        const set3 = set1.toggle(8);

        expect(set1.has(1, 2, 3, 4, 5, 6, 7)).toBeTruthy();
        expect(set2.has(1, 2, 3, 4, 5, 6)).toBeTruthy();
        expect(set2.has(7)).toBeFalsy();
        expect(set3.has(1, 2, 3, 4, 5, 6, 7, 8)).toBeTruthy();
    });

    it("should work with size", () => {
        const set = new FreezeTagSet([1, 2, 3, 4, 5, 6, 7]);
        expect(set.size()).toBe(7);
    });

    it("should work with toArray", () => {
        const test_arr = [1, 3, 4, 5, 6, 7, 8, 9, 10];
        const set1 = new FreezeTagSet([1, 2, 3, 4, 5, 6, 7]);
        const set2 = set1.add(8, 9, 10).remove(1).toggle(2, 1);

        const result_arr = set2.toArray();
        expect(result_arr.length).toBe(test_arr.length);

        for (const test_item of test_arr) {
            const contains =
                result_arr.filter((result_item) => result_item === test_item)
                    .length === 1;
            expect(contains).toBeTruthy();
        }
    });

    it("should be immutable on clone", () => {
        const set1 = new FreezeTagSet([1, 2, 3, 4]);
        const set2 = set1.clone();

        expect(Object.is(set1, set2)).toBeFalsy();
    });

    it("should be immutable on clear", () => {
        const set1 = new FreezeTagSet([1, 2, 3, 4, 5, 6, 7]);
        const set2 = set1.clear();

        expect(set1.has(1, 2, 3, 4, 5, 6, 7)).toBeTruthy();
        expect(set2.has(1, 2, 3, 4, 5, 6, 7)).toBeFalsy();
    });
});
