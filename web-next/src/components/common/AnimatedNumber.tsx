"use client";

import { useEffect, useRef, useState } from "react";
import { animate } from "motion/react";

type AnimatedNumberProps = {
  value: string | number | undefined;
  duration?: number;
};

export function AnimatedNumber({ value, duration = 800 }: AnimatedNumberProps) {
  const [displayValue, setDisplayValue] = useState(0);
  const prevValueRef = useRef(0);

  useEffect(() => {
    if (value === undefined || value === null || value === "-") {
      prevValueRef.current = 0;
      return;
    }

    const numericValue =
      typeof value === "string" ? parseFloat(value.replace(/,/g, "")) : value;

    if (Number.isNaN(numericValue)) {
      return;
    }

    const controls = animate(prevValueRef.current, numericValue, {
      duration: duration / 1000,
      ease: "easeOut",
      onUpdate: (latest) => {
        setDisplayValue(latest);
        prevValueRef.current = latest;
      },
    });

    return () => controls.stop();
  }, [value, duration]);

  if (value === undefined || value === null) {
    return <span>-</span>;
  }

  const shouldShowDecimals = typeof value === "string" && value.includes(".");
  const decimalPlaces = shouldShowDecimals ? 2 : 0;

  return (
    <span>
      {displayValue.toLocaleString("en-US", {
        minimumFractionDigits: decimalPlaces,
        maximumFractionDigits: decimalPlaces,
      })}
    </span>
  );
}
