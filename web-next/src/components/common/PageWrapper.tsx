"use client";

import { AnimatePresence, motion } from "motion/react";
import { Children, isValidElement, type ReactNode } from "react";
import { EASING } from "@/lib/animations/fluid-transitions";

type PageWrapperProps = {
  children: ReactNode;
  className?: string;
};

function getDiminishingDelay(index: number): number {
  if (index === 0) return 0;
  return Math.min(0.08 * Math.log2(index + 1), 0.4);
}

export function PageWrapper({ children, className = "space-y-6" }: PageWrapperProps) {
  const childArray = Children.toArray(children);

  return (
    <motion.div className={className}>
      <AnimatePresence>
        {childArray.map((child, index) => {
          const key = isValidElement(child) ? child.key : null;

          return (
            <motion.div
              key={key}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{
                opacity: 0,
                scale: 0.95,
                transition: { duration: 0.3 },
              }}
              transition={{
                duration: 0.5,
                ease: EASING.easeOutExpo,
                delay: getDiminishingDelay(index),
              }}
              layout
            >
              {child}
            </motion.div>
          );
        })}
      </AnimatePresence>
    </motion.div>
  );
}
