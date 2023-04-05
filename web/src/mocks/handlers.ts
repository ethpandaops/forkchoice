import { rest, RestHandler, MockedRequest, DefaultBodyType } from 'msw';

import {
  Response,
  V1GetEthereumNowResponse,
  V1GetEthereumSpecResponse,
  V1MetadataListResponse,
  V1MetadataListNodesResponse,
  V1GetFrameResponse,
} from '@app/types/api';
import { generateRandomForkChoiceData } from '@utils/api';
import { BASE_URL } from '@utils/environment';

export const networkName = 'goerli';

export const spec: Required<V1GetEthereumSpecResponse>['spec'] = {
  seconds_per_slot: 12,
  slots_per_epoch: 32,
  genesis_time: '2021-03-23T14:00:00Z',
};

export function getNow(): Required<V1GetEthereumNowResponse> {
  const slot = Math.floor(
    (Date.now() - new Date(spec.genesis_time).getTime()) / 1000 / spec.seconds_per_slot,
  );

  return {
    slot,
    epoch: Math.floor(slot / spec.slots_per_epoch),
  };
}

export const nodes: Required<V1MetadataListNodesResponse> = {
  nodes: ['ams3-teku-001', 'syd1-lighthouse-001', 'syd1-prysm-001'],
  pagination: { total: 3 },
};

export const handlers: RestHandler<MockedRequest<DefaultBodyType>>[] = [
  rest.get(`${BASE_URL}api/v1/ethereum/now`, async (_, res, ctx) => {
    return res(ctx.json({ data: getNow() }));
  }),
  rest.get(`${BASE_URL}api/v1/ethereum/spec`, (_, res, ctx) => {
    return res(ctx.json({ data: { network_name: networkName, spec } }));
  }),
  rest.post(`${BASE_URL}api/v1/metadata/nodes`, (_, res, ctx) => {
    return res(ctx.json({ data: nodes }));
  }),
  rest.post(`${BASE_URL}api/v1/metadata`, async (req, res, ctx) => {
    const { slot, epoch } = getNow();
    const data: Response<V1MetadataListResponse> = {
      data: {
        frames: [
          {
            id: 'bfe734bb-c986-4859-8b3e-44314ceca0b5',
            node: nodes.nodes[Math.floor(Math.random() * nodes.nodes.length)],
            fetched_at: new Date(
              new Date(spec.genesis_time).getTime() + slot * spec.seconds_per_slot * 1000,
            ).toISOString(),
            wall_clock_slot: slot,
            wall_clock_epoch: epoch,
            labels: [],
          },
        ],
        pagination: { total: 1 },
      },
    };
    return res(ctx.json(data));
  }),
  rest.get(`${BASE_URL}api/v1/frames/:id`, (req, res, ctx) => {
    const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
    const { slot, epoch } = getNow();
    const data: Response<V1GetFrameResponse> = {
      data: {
        frame: {
          data: generateRandomForkChoiceData(),
          metadata: {
            id,
            node: 'ams3-teku-001',
            fetched_at: new Date(
              new Date(spec.genesis_time).getTime() + slot * spec.seconds_per_slot * 1000,
            ).toISOString(),
            wall_clock_slot: slot,
            wall_clock_epoch: epoch,
            labels: [],
          },
        },
      },
    };
    return res(ctx.json(data));
  }),
];
