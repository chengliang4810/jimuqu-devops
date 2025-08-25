package com.jimuqu.devops.util;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

/**
 * 密码加密工具类
 */
public class PasswordUtil {
    
    private static final PasswordEncoder PASSWORD_ENCODER = new BCryptPasswordEncoder();
    
    /**
     * 加密密码
     */
    public static String encode(String rawPassword) {
        if (rawPassword == null || rawPassword.isEmpty()) {
            return null;
        }
        return PASSWORD_ENCODER.encode(rawPassword);
    }
    
    /**
     * 验证密码
     */
    public static boolean matches(String rawPassword, String encodedPassword) {
        if (rawPassword == null || encodedPassword == null) {
            return false;
        }
        return PASSWORD_ENCODER.matches(rawPassword, encodedPassword);
    }
}