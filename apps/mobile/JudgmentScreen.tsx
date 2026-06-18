import React, { useState } from "react";
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
  ActivityIndicator,
  Modal,
} from "react-native";
import { submitJudgment } from "@supperjumpin/api-client";

export type FactorScore = 1 | 2 | 3 | 4;

export interface JudgmentDetail {
  id: string;
  performerName: string;
  source: string;
  destination: string;
  food: string;
  caption: string;
  runningAverage: number;
  judgmentCount: number;
}

export interface JudgmentScores {
  transgression: FactorScore | null;
  creativity: FactorScore | null;
  commitment: FactorScore | null;
  presentation: FactorScore | null;
}

const FACTOR_ORDER: (keyof JudgmentScores)[] = [
  "transgression",
  "creativity",
  "commitment",
  "presentation",
];

const FACTOR_LABELS: Record<keyof JudgmentScores, string> = {
  transgression: "Transgression",
  creativity: "Creativity",
  commitment: "Commitment",
  presentation: "Presentation",
};

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

interface JudgmentScreenProps {
  jumpId: string;
  detail: JudgmentDetail;
  session?: { access_token: string; user: { id: string } };
  guestSessionId?: string;
  onSubmitted: () => void;
  onCancel: () => void;
  onSignIn?: () => void;
}

export default function JudgmentScreen({
  jumpId,
  detail,
  session,
  guestSessionId,
  onSubmitted,
  onCancel,
  onSignIn,
}: JudgmentScreenProps) {
  const [scores, setScores] = useState<JudgmentScores>({
    transgression: null,
    creativity: null,
    commitment: null,
    presentation: null,
  });
  const [showReceipt, setShowReceipt] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [guestCapped, setGuestCapped] = useState(false);

  const allSelected =
    scores.transgression !== null &&
    scores.creativity !== null &&
    scores.commitment !== null &&
    scores.presentation !== null;

  function setFactor(factor: keyof JudgmentScores, value: FactorScore) {
    setScores((prev) => ({ ...prev, [factor]: value }));
    setError(null);
  }

  async function handleFileJudgment() {
    if (!allSelected) return;
    setSubmitting(true);
    setError(null);
    setGuestCapped(false);
    try {
      await submitJudgment({
        baseUrl: API_BASE,
        accessToken: session?.access_token,
        guestSessionId,
        jumpId,
        commitment: scores.commitment!,
        transgression: scores.transgression!,
        creativity: scores.creativity!,
        presentation: scores.presentation!,
      });
      onSubmitted();
    } catch (e: any) {
      const msg = e.message ?? "Could not file Judgment";
      if (msg.toLowerCase().includes("guest judgment cap reached") || msg.toLowerCase().includes("guest_cap")) {
        setGuestCapped(true);
      } else {
        setError(msg);
      }
      setSubmitting(false);
    }
  }

  return (
    <View style={styles.screen}>
      <View style={styles.header}>
        <TouchableOpacity onPress={onCancel} accessibilityRole="button" accessibilityLabel="Cancel Judgment">
          <Text style={styles.cancelText}>Cancel</Text>
        </TouchableOpacity>
        <Text style={styles.title}>Judge this Jump</Text>
        <View style={styles.headerSpacer} />
      </View>

      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.jumpSummary}>
          <Text style={styles.food}>{detail.food}</Text>
          <Text style={styles.route}>
            {detail.source} → {detail.destination}
          </Text>
          <Text style={styles.caption}>{detail.caption}</Text>
        </View>

        {FACTOR_ORDER.map((factor) => {
          const factorSelected = scores[factor] !== null;
          return (
            <View
              key={factor}
              style={styles.factorSection}
              accessible
              accessibilityLabel={`${FACTOR_LABELS[factor]} factor, ${factorSelected ? 'selected' : 'not selected'}`}
            >
              <Text style={styles.factorLabel}>{FACTOR_LABELS[factor]}</Text>
              <View style={styles.factorButtons}>
                {[1, 2, 3, 4].map((value) => {
                  const selected = scores[factor] === value;
                  return (
                    <TouchableOpacity
                      key={value}
                      style={[
                        styles.factorButton,
                        selected && styles.factorButtonSelected,
                      ]}
                      onPress={() => setFactor(factor, value as FactorScore)}
                      accessibilityRole="button"
                      accessibilityLabel={`${FACTOR_LABELS[factor]} score ${value}${selected ? ', selected' : ''}`}
                      accessibilityState={{ selected }}
                    >
                      <Text
                        style={[
                          styles.factorButtonText,
                          selected && styles.factorButtonTextSelected,
                        ]}
                      >
                        {value}
                      </Text>
                    </TouchableOpacity>
                  );
                })}
              </View>
            </View>
          );
        })}

        <TouchableOpacity
          style={[styles.reviewButton, !allSelected && styles.reviewButtonDisabled]}
          disabled={!allSelected}
          onPress={() => allSelected && setShowReceipt(true)}
          accessibilityRole="button"
          accessibilityLabel={allSelected ? "Review Judgment, enabled" : "Review Judgment, disabled"}
          accessibilityState={{ disabled: !allSelected }}
        >
          <Text style={styles.reviewButtonText}>Review Judgment</Text>
        </TouchableOpacity>
      </ScrollView>

      <Modal
        visible={showReceipt}
        transparent
        animationType="slide"
        onRequestClose={() => !submitting && setShowReceipt(false)}
        accessibilityLabel="Judgment Receipt"
      >
        <View style={styles.receiptBackdrop}>
          <View style={styles.receiptSheet}>
            <Text style={styles.receiptTitle}>Judgment Receipt</Text>
            <Text style={styles.receiptSubtitle}>Review your verdict before filing.</Text>

            {FACTOR_ORDER.map((factor) => (
              <View
                key={factor}
                style={styles.receiptRow}
                accessibilityLabel={`Judgment Receipt factor ${FACTOR_LABELS[factor]} score ${scores[factor]}`}
              >
                <Text style={styles.receiptFactor}>{FACTOR_LABELS[factor]}</Text>
                <Text style={styles.receiptScore}>{scores[factor]}</Text>
              </View>
            ))}

            {guestCapped && (
              <View style={styles.capPrompt} accessible accessibilityLabel="Guest cap reached prompt">
                <Text style={styles.capTitle}>You've reached the guest judging limit</Text>
                <Text style={styles.capBody}>Create an account to keep judging and save your history.</Text>
                {onSignIn && (
                  <TouchableOpacity
                    style={styles.signInButton}
                    onPress={onSignIn}
                    accessibilityRole="button"
                    accessibilityLabel="Sign In to Keep Judging"
                  >
                    <Text style={styles.signInButtonText}>Sign In to Keep Judging</Text>
                  </TouchableOpacity>
                )}
              </View>
            )}

            {!guestCapped && error && (
              <View style={styles.errorRow} accessibilityLabel="Judgment submission error">
                <Text style={styles.errorText}>{error}</Text>
              </View>
            )}

            {!guestCapped && (
              <TouchableOpacity
                style={[styles.fileButton, submitting && styles.fileButtonDisabled]}
                disabled={submitting}
                onPress={handleFileJudgment}
                accessibilityRole="button"
                accessibilityLabel="File Judgment"
                accessibilityState={{ disabled: submitting }}
              >
                {submitting ? (
                  <ActivityIndicator size="small" color="#fffaf2" />
                ) : (
                  <Text style={styles.fileButtonText}>File Judgment</Text>
                )}
              </TouchableOpacity>
            )}

            <TouchableOpacity
              style={styles.editButton}
              disabled={submitting}
              onPress={() => setShowReceipt(false)}
              accessibilityRole="button"
              accessibilityLabel="Edit Judgment"
              accessibilityState={{ disabled: submitting }}
            >
              <Text style={styles.editButtonText}>Edit</Text>
            </TouchableOpacity>
          </View>
        </View>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    paddingVertical: 12,
    minHeight: 56,
  },
  cancelText: {
    color: "#c1673a",
    fontSize: 16,
    fontWeight: "700",
  },
  title: {
    color: "#2f241d",
    fontSize: 18,
    fontWeight: "800",
  },
  headerSpacer: {
    width: 50,
  },
  content: {
    padding: 16,
    gap: 20,
  },
  jumpSummary: {
    backgroundColor: "#fffaf2",
    borderRadius: 16,
    padding: 16,
    borderWidth: 2,
    borderColor: "#2f241d",
    gap: 4,
  },
  food: {
    color: "#2f241d",
    fontSize: 20,
    fontWeight: "900",
  },
  route: {
    color: "#c1673a",
    fontSize: 14,
    fontWeight: "700",
  },
  caption: {
    color: "#4d3b31",
    fontSize: 14,
    marginTop: 8,
  },
  factorSection: {
    gap: 8,
  },
  factorLabel: {
    color: "#2f241d",
    fontSize: 16,
    fontWeight: "800",
  },
  factorButtons: {
    flexDirection: "row",
    gap: 8,
  },
  factorButton: {
    flex: 1,
    backgroundColor: "#fffaf2",
    borderWidth: 2,
    borderColor: "#2f241d",
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: "center",
    minHeight: 48,
  },
  factorButtonSelected: {
    backgroundColor: "#2f241d",
    borderColor: "#2f241d",
  },
  factorButtonText: {
    color: "#2f241d",
    fontSize: 16,
    fontWeight: "800",
  },
  factorButtonTextSelected: {
    color: "#fffaf2",
  },
  reviewButton: {
    backgroundColor: "#c1673a",
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: "center",
    marginTop: 8,
    minHeight: 48,
  },
  reviewButtonDisabled: {
    backgroundColor: "#e8ddd2",
  },
  reviewButtonText: {
    color: "#fffaf2",
    fontSize: 16,
    fontWeight: "800",
  },
  receiptBackdrop: {
    flex: 1,
    justifyContent: "flex-end",
    backgroundColor: "rgba(47, 36, 29, 0.5)",
  },
  receiptSheet: {
    backgroundColor: "#fffaf2",
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    padding: 24,
    gap: 12,
  },
  receiptTitle: {
    color: "#2f241d",
    fontSize: 22,
    fontWeight: "900",
  },
  receiptSubtitle: {
    color: "#4d3b31",
    fontSize: 14,
    marginBottom: 8,
  },
  receiptRow: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingVertical: 10,
    borderBottomWidth: 1,
    borderBottomColor: "#e8ddd2",
  },
  receiptFactor: {
    color: "#2f241d",
    fontSize: 16,
    fontWeight: "700",
  },
  receiptScore: {
    color: "#c1673a",
    fontSize: 18,
    fontWeight: "900",
  },
  errorRow: {
    backgroundColor: "#f9e3d4",
    borderRadius: 8,
    padding: 12,
  },
  errorText: {
    color: "#c1673a",
    fontSize: 14,
    fontWeight: "700",
  },
  fileButton: {
    backgroundColor: "#c1673a",
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: "center",
    minHeight: 48,
    marginTop: 8,
  },
  fileButtonDisabled: {
    backgroundColor: "#e8ddd2",
  },
  fileButtonText: {
    color: "#fffaf2",
    fontSize: 16,
    fontWeight: "800",
  },
  editButton: {
    paddingVertical: 12,
    alignItems: "center",
    minHeight: 44,
  },
  editButtonText: {
    color: "#4d3b31",
    fontSize: 16,
    fontWeight: "700",
  },
  capPrompt: {
    backgroundColor: "#f9e3d4",
    borderRadius: 12,
    padding: 16,
    gap: 8,
  },
  capTitle: {
    color: "#2f241d",
    fontSize: 16,
    fontWeight: "800",
  },
  capBody: {
    color: "#4d3b31",
    fontSize: 14,
  },
  signInButton: {
    backgroundColor: "#c1673a",
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: "center",
    minHeight: 48,
    marginTop: 4,
  },
  signInButtonText: {
    color: "#fffaf2",
    fontSize: 16,
    fontWeight: "800",
  },
});
