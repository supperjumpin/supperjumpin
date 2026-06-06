import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, ActivityIndicator } from 'react-native';

interface LoginScreenProps {
  onSignIn: () => Promise<void>;
  loading?: boolean;
}

export default function LoginScreen({ onSignIn, loading }: LoginScreenProps) {
  return (
    <View style={styles.container} accessibilityLabel="Sign in screen">
      <Text style={styles.title}>Supperjumpin</Text>
      <Text style={styles.subtitle}>Sign in to start jumping</Text>
      <TouchableOpacity
        style={styles.button}
        onPress={onSignIn}
        disabled={loading}
        accessibilityRole="button"
        accessibilityLabel="Sign In"
      >
        {loading ? (
          <ActivityIndicator color="#fffaf2" />
        ) : (
          <Text style={styles.buttonText}>Sign In</Text>
        )}
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
    backgroundColor: '#f7efe2',
  },
  title: {
    fontSize: 36,
    fontWeight: '900',
    color: '#2f241d',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#4d3b31',
    marginBottom: 32,
  },
  button: {
    backgroundColor: '#c1673a',
    paddingHorizontal: 32,
    paddingVertical: 14,
    borderRadius: 12,
    minHeight: 48,
    justifyContent: 'center',
    alignItems: 'center',
  },
  buttonText: {
    color: '#fffaf2',
    fontSize: 16,
    fontWeight: '700',
  },
});