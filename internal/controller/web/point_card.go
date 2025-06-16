package web

import (
	"errors"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/controller/web/util"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/utils/errs"
	"strconv"
)

func ProvidePointCardController(srv interfaces.IPointCardSrv, reqParse util.IRequestParse, logger logger.ILogger) *PointCardCtrl {
	return &PointCardCtrl{
		srv:      srv,
		reqParse: reqParse,
		logger:   logger,
	}
}

type PointCardCtrl struct {
	srv      interfaces.IPointCardSrv
	reqParse util.IRequestParse
	logger   logger.ILogger
}

func (ctrl *PointCardCtrl) KintonePointCardWebhook(ctx *gin.Context) {
	req := dto.KintoneWebhookPointCardIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		ctrl.logger.Error(ctx, "studentCtrl KintonePointCardWebhook Bind", err)
		return
	}

	switch req.Type {
	case kintone.AddRecordType:
		ctrl.addPointCard(ctx, req)
	case kintone.UpdateRecordType, kintone.UpdateStatusType:
		ctrl.updatePointCard(ctx, req)
	case kintone.DeleteRecordType:
		ctrl.deletePointCard(ctx, req)
	}
}

func (ctrl *PointCardCtrl) addPointCard(ctx *gin.Context, req dto.KintoneWebhookPointCardIO) {
	boPointCard, err := ctrl.checkPointCardBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "PointCardCtrl addPointCard checkPointCardBasicRequestData", err, zap.String("record_ref_id", req.Record.Id.Value))
		return
	}

	err = ctrl.srv.AddPointCard(ctx, boPointCard, &bo.StudentCond{StudentName: boPointCard.StudentName, ParentPhone: boPointCard.ParentPhone})
	if err != nil {
		ctrl.logger.Error(ctx, "PointCardCtrl addPointCard srv.AddPointCard", err, zap.Int("record_ref_id", boPointCard.RecordRefId))
	}
}

func (ctrl *PointCardCtrl) updatePointCard(ctx *gin.Context, req dto.KintoneWebhookPointCardIO) {
	boPointCard, err := ctrl.checkPointCardBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "PointCardCtrl updatePointCard checkPointCardBasicRequestData", err, zap.String("record_ref_id", req.Record.Id.Value))
		return
	}

	err = ctrl.srv.UpdatePointCard(ctx,
		&bo.UpdatePointCardCond{RecordRefId: boPointCard.RecordRefId},
		&bo.StudentCond{StudentName: boPointCard.StudentName, ParentPhone: boPointCard.ParentPhone},
		&bo.UpdatePointCardRecordData{RestPoints: &boPointCard.RestPoints},
	)
	if err != nil {
		ctrl.logger.Error(ctx, "PointCardCtrl updatePointCard srv.UpdatePointCard", err, zap.Int("record_ref_id", boPointCard.RecordRefId))
	}
}

func (ctrl *PointCardCtrl) deletePointCard(ctx *gin.Context, req dto.KintoneWebhookPointCardIO) {
	recordRefId, err := strconv.Atoi(req.RecordId)
	if err != nil {
		ctrl.logger.Error(ctx, "PointCardCtrl deletePointCard Atoi", err, zap.String("record_id", req.RecordId))
		return
	}

	if err = ctrl.srv.DeletePointCard(ctx, &bo.UpdatePointCardCond{RecordRefId: recordRefId}); err != nil {
		ctrl.logger.Error(ctx, "PointCardCtrl deletePointCard DeletePointCard", err, zap.Int("record_ref_id", recordRefId))
	}
}

func (ctrl *PointCardCtrl) checkPointCardBasicRequestData(req dto.KintoneWebhookPointCardIO) (*bo.PointCard, error) {
	if req.Record.Id.Value == "" {
		return nil, errors.New("invalid record id")
	}

	if req.Record.StudentName.Value == "" {
		return nil, errs.StudentErr.StudentNameInvalidErr
	}

	boPointCard, err := req.Record.ToPointCard()
	if err != nil {
		return nil, xerrors.Errorf("req.Record.ToPointCard: %w", err)
	}

	return boPointCard, nil
}
