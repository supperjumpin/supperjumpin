import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';

interface CreateJumpScreenProps {
  onBack: () => void;
}

export default function CreateJumpScreen({ onBack }: CreateJumpScreenProps) {
  return (
    <View style={styles.container} accessibilityLabel="Post Jump screen">
      <Text style={styles.title}>Post a Jump</Text>
      <Text style={styles.subtitle}>Coming soon — select your Source, Destination, and Food, then capture evidence.</Text>
      <TouchableOpacity
        style={styles.backButton}
        onPress={onBack}
        accessibilityRole="button"
        accessibilityLabel="Back to feed"
      >
        <Text style={styles.backButtonText}>Back to Feed</Text>
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
    fontSize: 28,
    fontWeight: '900',
    color: '#2f241d',
    marginBottom: 16,
  },
  subtitle: {
    fontSize: 16,
    color: '#4d3b31',
    textAlign: 'center',
    lineHeight: 22,
    marginBottom: 32,
  },
  backButton: {
    backgroundColor: '#c1673a',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 12,
    minHeight: 48,
    justifyContent: 'center',
  },
  backButtonText: {
    color: '#fffaf2',
    fontSize: 16,
    fontWeight: '700',
  },
});