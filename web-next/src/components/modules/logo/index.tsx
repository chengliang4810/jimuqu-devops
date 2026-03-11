'use client';

import { motion } from 'motion/react';

interface LogoProps {
    size?: number | string;
}

export default function Logo({ size = 48 }: LogoProps) {
    const sizeValue = size === '100%' ? '100%' : size;

    return (
        <motion.svg
            viewBox="0 0 100 100"
            xmlns="http://www.w3.org/2000/svg"
            width={sizeValue}
            height={sizeValue}
            className="text-primary"
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.3 }}
        >
            {/* 积木堆叠形状 - 简化的 Logo */}
            <path
                d="M50 15 L85 40 L85 70 L70 85 L30 85 L15 70 L15 40 Z"
                fill="none"
                stroke="currentColor"
                strokeWidth="6"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
            <path
                d="M35 45 L50 35 L65 45 L65 65 L50 75 L35 65 Z"
                fill="none"
                stroke="currentColor"
                strokeWidth="5"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
            <path
                d="M45 55 L50 50 L55 55 L55 65 L50 70 L45 65 Z"
                fill="none"
                stroke="currentColor"
                strokeWidth="4"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </motion.svg>
    );
}
