/**
 * Copyright 2026 Flant JSC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useEffect, useState } from 'react';
import {
  Box,
  ButtonPrimary,
  ButtonSecondary,
  Input,
  LabelInput,
  Text,
} from 'design';
import { Danger } from 'design/Alert';
import Dialog, {
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from 'design/DialogConfirmation';
import Select, { Option } from 'shared/components/Select';

import cfg from 'teleport/config';
import {
  createAccessRequest,
  getDurationOptions,
} from 'teleport/AccessRequests/service';
import { AccessRequest, DurationOption } from 'teleport/AccessRequests/types';
import { AccessStrategy } from 'teleport/services/user/types';

type Props = {
  requestableRoles: string[];
  suggestedReviewers: string[];
  accessStrategy: AccessStrategy;
  onClose: () => void;
  onCreated: (req: AccessRequest) => void;
};

export function NewAccessRequestDialog({
  requestableRoles,
  suggestedReviewers,
  accessStrategy,
  onClose,
  onCreated,
}: Props) {
  const reasonRequired = accessStrategy.type === 'reason';
  const [selectedRoles, setSelectedRoles] = useState<Option[]>([]);
  const [reason, setReason] = useState('');
  const [reviewersInput, setReviewersInput] = useState(
    suggestedReviewers.join(', ')
  );
  const [startTime, setStartTime] = useState('');
  const [durationOptions, setDurationOptions] = useState<DurationOption[]>([]);
  const [selectedDuration, setSelectedDuration] =
    useState<DurationOption | null>(null);
  const [durationStatus, setDurationStatus] = useState<
    'idle' | 'loading' | 'error'
  >('idle');
  const [status, setStatus] = useState<'idle' | 'processing'>('idle');
  const [error, setError] = useState('');

  const roleOptions: Option[] = requestableRoles.map(role => ({
    label: role,
    value: role,
  }));

  useEffect(() => {
    let cancelled = false;
    if (selectedRoles.length === 0) {
      setDurationOptions([]);
      setSelectedDuration(null);
      setDurationStatus('idle');
      return;
    }
    setDurationStatus('loading');
    getDurationOptions(
      cfg.proxyCluster,
      selectedRoles.map(o => o.value),
      []
    )
      .then(opts => {
        if (cancelled) return;
        setDurationOptions(opts);
        setSelectedDuration(opts[opts.length - 1] || null);
        setDurationStatus('idle');
      })
      .catch(() => {
        if (cancelled) return;
        setDurationOptions([]);
        setSelectedDuration(null);
        setDurationStatus('error');
      });
    return () => {
      cancelled = true;
    };
  }, [selectedRoles]);

  const isSubmitDisabled =
    status === 'processing' ||
    selectedRoles.length === 0 ||
    (reasonRequired && reason.trim() === '');

  async function handleSubmit() {
    setStatus('processing');
    setError('');
    try {
      const maxDuration = selectedDuration
        ? new Date(selectedDuration.value)
        : undefined;
      const assumeStartTime = startTime ? new Date(startTime) : undefined;
      if (assumeStartTime && Number.isNaN(assumeStartTime.getTime())) {
        throw new Error('Invalid start time');
      }
      const reviewers = reviewersInput
        .split(',')
        .map(s => s.trim())
        .filter(Boolean);
      const req = await createAccessRequest(
        cfg.proxyCluster,
        selectedRoles.map(o => o.value),
        [],
        reason,
        false,
        maxDuration,
        assumeStartTime,
        reviewers
      );
      onCreated(req);
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to create access request');
      setStatus('idle');
    }
  }

  return (
    <Dialog disableEscapeKeyDown={false} onClose={onClose} open={true}>
      <DialogHeader>
        <DialogTitle>New Access Request</DialogTitle>
      </DialogHeader>
      <DialogContent width="540px">
        {error && <Danger>{error}</Danger>}

        {requestableRoles.length === 0 ? (
          <Text typography="paragraph" mb="3">
            You do not have any requestable roles. Ask your administrator to add
            roles to your{' '}
            <Text as="span" bold>
              allow.request.roles
            </Text>{' '}
            field.
          </Text>
        ) : (
          <>
            <Box mb={3}>
              <LabelInput>Roles to Request</LabelInput>
              <Select
                isMulti
                placeholder="Select one or more roles"
                value={selectedRoles}
                options={roleOptions}
                onChange={(opts: Option[] | null) =>
                  setSelectedRoles(opts || [])
                }
              />
            </Box>

            <Box mb={3}>
              <LabelInput>Max Duration</LabelInput>
              <Select
                isDisabled={
                  selectedRoles.length === 0 ||
                  durationStatus === 'loading' ||
                  durationOptions.length === 0
                }
                placeholder={
                  durationStatus === 'loading'
                    ? 'Loading...'
                    : durationStatus === 'error'
                      ? 'Failed to load durations'
                      : selectedRoles.length === 0
                        ? 'Select roles first'
                        : 'Select max duration'
                }
                value={selectedDuration}
                options={durationOptions}
                onChange={(opt: DurationOption | null) =>
                  setSelectedDuration(opt)
                }
                isClearable={false}
              />
              <Text fontSize={0} color="text.slightlyMuted" mt={1}>
                Max allowed by your role's <code>request.max_duration</code>.
              </Text>
            </Box>

            <Box mb={3}>
              <LabelInput>Start Time (optional)</LabelInput>
              <Input
                type="datetime-local"
                value={startTime}
                onChange={e => setStartTime(e.target.value)}
              />
              <Text fontSize={0} color="text.slightlyMuted" mt={1}>
                When the approved roles can be assumed. Leave empty to assume
                immediately after approval.
              </Text>
            </Box>

            <Box mb={3}>
              <LabelInput>
                Reason{reasonRequired ? ' (required)' : ' (optional)'}
              </LabelInput>
              <Input
                placeholder={
                  accessStrategy.prompt || 'Describe why you need this access'
                }
                value={reason}
                onChange={e => setReason(e.target.value)}
              />
            </Box>

            <Box>
              <LabelInput>Suggested Reviewers (optional)</LabelInput>
              <Input
                placeholder="alice, bob"
                value={reviewersInput}
                onChange={e => setReviewersInput(e.target.value)}
              />
              <Text fontSize={0} color="text.slightlyMuted" mt={1}>
                Comma-separated usernames. Only a hint for notifications and the
                request UI — does not grant approval rights. If left empty,
                defaults from your role's{' '}
                <code>allow.request.suggested_reviewers</code> are used.
              </Text>
            </Box>
          </>
        )}
      </DialogContent>
      <DialogFooter>
        <ButtonPrimary
          mr="3"
          disabled={isSubmitDisabled}
          onClick={handleSubmit}
        >
          {status === 'processing' ? 'Creating...' : 'Create Request'}
        </ButtonPrimary>
        <ButtonSecondary disabled={status === 'processing'} onClick={onClose}>
          Cancel
        </ButtonSecondary>
      </DialogFooter>
    </Dialog>
  );
}
