import React, { useRef, useEffect } from "react";

export const enum CheckboxState {
    Checked = 0,
    Unchecked = 1,
    Indeterminate = -1,
}

export type IndeterminateCheckboxProps = {
    value: CheckboxState;
};

export default function IndeterminateCheckbox({
    value,
    ...otherProps
}: IndeterminateCheckboxProps) {
    const checkRef = useRef<HTMLInputElement | null>(null);

    useEffect(() => {
        if (checkRef.current == null) {
            console.error("checkRef should not be null!");
            return;
        }
        checkRef.current.checked = value === CheckboxState.Checked;
        checkRef.current.indeterminate = value === CheckboxState.Indeterminate;
    }, [value]);

    return <input type="checkbox" ref={checkRef} {...otherProps} />;
}
