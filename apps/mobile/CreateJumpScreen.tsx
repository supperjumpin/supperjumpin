import React, { useState } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, TextInput, ScrollView } from 'react-native';
import { createJump } from '@supperjumpin/api-client';

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? 'http://localhost:8080';

function isStorageObjectKey(value: string): boolean {
  const trimmed = value.trim();
  return Boolean(trimmed) && !/^(https?:|file:|content:)/i.test(trimmed);
}

interface CreateJumpScreenProps {
  session: { access_token: string; user: { id: string } } | null;
  onBack: () => void;
}

export default function CreateJumpScreen({ session, onBack }: CreateJumpScreenProps) {
  const [source, setSource] = useState('');
  const [destination, setDestination] = useState('');
  const [food, setFood] = useState('');
  const [caption, setCaption] = useState('');
  const [mediaObjectKey, setMediaObjectKey] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canSubmit = Boolean(
    session?.access_token && source.trim() && destination.trim() && food.trim() && caption.trim() && isStorageObjectKey(mediaObjectKey),
  );

  const handleSubmit = async () => {
    if (!session?.access_token || !canSubmit) return;

    setSubmitting(true);
    setError(null);
    try {
      await createJump({
        baseUrl: API_BASE,
        accessToken: session.access_token,
        source: source.trim(),
        destination: destination.trim(),
        food: food.trim(),
        caption: caption.trim(),
        mediaObjectKey: mediaObjectKey.trim(),
      });
      onBack();
    } catch (err: any) {
      setError(err?.message ?? 'Could not create Jump');
      setSubmitting(false);
    }
  };

  return (
    <View style={styles.container} accessibilityLabel="Post Jump screen">
      <ScrollView
        testID="create-jump-scroll"
        keyboardShouldPersistTaps="handled"
        contentContainerStyle={styles.content}
      >
        <Text style={styles.title}>Post a Jump</Text>
        <Text style={styles.subtitle}>Capture one evidence photo, then fill out the route and caption.</Text>

        <TextInput
          style={styles.input}
          value={source}
          onChangeText={setSource}
          placeholder="Source"
          accessibilityLabel="Source"
          testID="source-input"
        />
        <TextInput
          style={styles.input}
          value={destination}
          onChangeText={setDestination}
          placeholder="Destination"
          accessibilityLabel="Destination"
          testID="destination-input"
        />
        <TextInput
          style={styles.input}
          value={food}
          onChangeText={setFood}
          placeholder="Food"
          accessibilityLabel="Food"
          testID="food-input"
        />
        <TextInput
          style={[styles.input, styles.multilineInput]}
          value={caption}
          onChangeText={setCaption}
          placeholder="Caption"
          multiline
          accessibilityLabel="Caption"
          testID="caption-input"
        />
        <TextInput
          style={styles.input}
          value={mediaObjectKey}
          onChangeText={setMediaObjectKey}
          placeholder="Paste uploaded evidence object key"
          accessibilityLabel="Evidence photo"
          autoCapitalize="none"
          testID="evidence-photo-input"
        />

        {error ? <Text style={styles.errorText} accessibilityLabel={error}>{error}</Text> : null}

        <TouchableOpacity
          style={[styles.submitButton, (!canSubmit || submitting) && styles.submitButtonDisabled]}
          onPress={handleSubmit}
          disabled={!canSubmit || submitting}
          accessibilityState={{ disabled: !canSubmit || submitting }}
          accessibilityRole="button"
          accessibilityLabel="Submit Jump"
          testID="submit-jump-button"
        >
          <Text style={styles.submitButtonText}>{submitting ? 'Submitting...' : 'Submit Jump'}</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.backButton}
          onPress={onBack}
          accessibilityRole="button"
          accessibilityLabel="Back to feed"
        >
          <Text style={styles.backButtonText}>Back to Feed</Text>
        </TouchableOpacity>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f7efe2',
  },
  content: {
    flexGrow: 1,
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
    marginBottom: 20,
  },
  input: {
    width: '100%',
    backgroundColor: '#fffaf2',
    borderColor: '#2f241d',
    borderWidth: 2,
    borderRadius: 12,
    minHeight: 48,
    paddingHorizontal: 12,
    paddingVertical: 10,
    marginBottom: 12,
    color: '#2f241d',
  },
  multilineInput: {
    minHeight: 96,
    textAlignVertical: 'top',
  },
  errorText: {
    color: '#b63d20',
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 12,
  },
  submitButton: {
    backgroundColor: '#2f241d',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 12,
    minHeight: 48,
    justifyContent: 'center',
    width: '100%',
    marginBottom: 12,
  },
  submitButtonDisabled: {
    backgroundColor: '#8c7a6f',
  },
  submitButtonText: {
    color: '#fffaf2',
    fontSize: 16,
    fontWeight: '700',
    textAlign: 'center',
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
