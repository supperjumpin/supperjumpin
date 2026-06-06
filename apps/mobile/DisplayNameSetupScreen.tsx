import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';

interface DisplayNameSetupScreenProps {
  onSubmit: (displayName: string) => Promise<void>;
}

export default function DisplayNameSetupScreen({ onSubmit }: DisplayNameSetupScreenProps) {
  const [name, setName] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSave = async () => {
    const trimmed = name.trim();
    if (!trimmed) {
      setError('Display name is required');
      return;
    }
    setError(null);
    setLoading(true);
    try {
      await onSubmit(trimmed);
    } catch {
      setError('Could not save display name');
    } finally {
      setLoading(false);
    }
  };

  return (
    <View style={styles.container} accessibilityLabel="Display name setup screen">
      <Text style={styles.title}>Choose your display name</Text>
      <Text style={styles.subtitle}>
        This is how other players will see you
      </Text>
      <TextInput
        style={styles.input}
        placeholder="Enter your display name"
        placeholderTextColor="#888"
        value={name}
        onChangeText={(text) => {
          setName(text);
          if (error) setError(null);
        }}
        accessibilityLabel="Display name input"
      />
      {error && (
        <Text style={styles.error} accessibilityLabel={error}>
          {error}
        </Text>
      )}
      <TouchableOpacity
        style={styles.button}
        onPress={handleSave}
        disabled={loading}
        accessibilityRole="button"
        accessibilityLabel="Save display name"
      >
        {loading ? (
          <ActivityIndicator color="#fffaf2" />
        ) : (
          <Text style={styles.buttonText}>Save</Text>
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
    fontSize: 24,
    fontWeight: '800',
    color: '#2f241d',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 14,
    color: '#4d3b31',
    marginBottom: 24,
  },
  input: {
    width: '100%',
    backgroundColor: '#fffaf2',
    borderColor: '#2f241d',
    borderWidth: 2,
    borderRadius: 12,
    padding: 14,
    fontSize: 16,
    color: '#2f241d',
    marginBottom: 16,
  },
  error: {
    color: '#c1673a',
    fontSize: 14,
    marginBottom: 12,
  },
  button: {
    backgroundColor: '#c1673a',
    paddingHorizontal: 32,
    paddingVertical: 14,
    borderRadius: 12,
    minHeight: 48,
    justifyContent: 'center',
    alignItems: 'center',
    width: '100%',
  },
  buttonText: {
    color: '#fffaf2',
    fontSize: 16,
    fontWeight: '700',
  },
});